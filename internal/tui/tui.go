package tui

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/bogdannbv/armsync/internal/config"
	"github.com/bogdannbv/armsync/internal/manager"
	"github.com/bogdannbv/armsync/internal/steam"
	"github.com/bogdannbv/armsync/pkg/reforger/auth"
	"github.com/bogdannbv/armsync/pkg/reforger/game"
	"github.com/bogdannbv/armsync/pkg/reforger/workshop"
)

type Model struct {
	keys *KeyMap

	ctx context.Context

	sizes sizes

	game    *game.Client
	manager *manager.Manager

	servers list.Model
	addons  list.Model
	bar     progress.Model

	install chan tea.Msg

	selected *game.Server

	fatal error
}

func NewModel(ctx context.Context) Model {
	sl := list.New(
		[]list.Item{},
		list.NewDefaultDelegate(),
		0,
		0,
	)
	sl.DisableQuitKeybindings()
	sl.Title = "Favorite Servers"

	ml := list.New(
		[]list.Item{},
		list.NewDefaultDelegate(),
		0, 0,
	)
	ml.SetShowStatusBar(false)
	ml.DisableQuitKeybindings()
	ml.Title = "Addons"

	serverKeys := func() []key.Binding { return []key.Binding{Keys.Enter, Keys.Refresh, Keys.Quit} }
	addonKeys := func() []key.Binding { return []key.Binding{Keys.Back, Keys.Quit} }
	sl.AdditionalFullHelpKeys = serverKeys
	sl.AdditionalShortHelpKeys = serverKeys
	ml.AdditionalFullHelpKeys = addonKeys
	ml.AdditionalShortHelpKeys = addonKeys

	return Model{
		ctx: ctx,

		keys: Keys,

		game:    nil,
		manager: nil,

		servers: sl,
		addons:  ml,
		bar:     progress.New(progress.WithDefaultBlend()),

		fatal: nil,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initDeps)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// handle only the most necessary messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		styles.ScreenBorder = styles.ScreenBorder.Width(msg.Width).Height(msg.Height)
		m.sizes = m.sizes.SetScreenSize(
			msg.Width-styles.ScreenBorder.GetHorizontalBorderSize(),
			msg.Height-styles.ScreenBorder.GetVerticalFrameSize(),
		)

		m.servers.SetSize(m.sizes.GetContentSize())
		m.addons.SetSize(m.sizes.GetContentSize())
		w, _ := m.sizes.GetContentSize()
		m.bar.SetWidth(w - 4)
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	}

	// unrecoverable state
	if m.fatal != nil {
		return m, nil
	}

	var (
		cmd        tea.Cmd
		serversCmd tea.Cmd
		addonsCmd  tea.Cmd
		cmds       []tea.Cmd
	)

	switch msg := msg.(type) {
	case initMsg:
		log.Debug("got init message")
		m.game = msg.Game
		m.manager = msg.Manager
		cmd = m.fetchFavServers()
	case favServersMsg:
		log.Debug("got fav servers message")
		sv := make([]list.Item, 0, len(msg.Servers))
		for _, s := range msg.Servers {
			sv = append(sv, server{Server: s})
		}
		serversCmd = m.servers.SetItems(sv)
	case fatalMsg:
		m.fatal = msg
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			if s := m.servers.SelectedItem(); m.selected == nil && s != nil {
				switch s := s.(type) {
				case server:
					log.Debug("selected server", "id", s.ID)
					cmd = selectServer(&s)
					//m.selected = &s.Server
				}
			}
		case key.Matches(msg, m.keys.Refresh):
			if m.selected == nil && m.game != nil && m.servers.FilterState() != list.Filtering {
				cmd = m.fetchFavServers()
			}
		case key.Matches(msg, m.keys.Back):
			if m.selected != nil && m.install == nil {
				m.selected = nil
			}
		}
	case serverSelectedMsg:
		m.selected = msg.Server
		diff := m.manager.DiffInstalled(m.selected.Mods)
		ads := make([]list.Item, 0, len(diff.ToInstall)+len(diff.ToUpdate))
		for _, mod := range diff.ToInstall {
			ads = append(ads, addon{
				id:     mod.ModID,
				name:   mod.Name,
				target: mod.Version,
			})
		}
		for _, ad := range diff.ToUpdate {
			ads = append(ads, addon{
				id:      ad.ID,
				name:    ad.Name,
				version: ad.Version,
				target:  ad.TargetVersion,
			})
		}
		slices.SortFunc(ads, func(a, b list.Item) int {
			return cmp.Compare(a.(addon).name, b.(addon).name)
		})
		m.addons.ResetSelected()
		m.addons.ResetFilter()
		cmd = m.addons.SetItems(ads)
		m.setTitle()
		if mods := m.pendingMods(); len(mods) > 0 {
			m.install = make(chan tea.Msg)
			cmd = tea.Batch(cmd, m.installMods(mods, m.install), waitInstall(m.install))
		} else {
			cmd = tea.Batch(cmd, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return backMsg{} }))
		}
	case backMsg:
		if m.install == nil {
			m.selected = nil
		}
	case installEventMsg:
		cmd = tea.Batch(m.updateAddon(manager.Event(msg)), waitInstall(m.install))
	case installDoneMsg:
		m.install = nil
		cmd = m.finishInstall(msg.err)
	}

	// Once a server is selected don't update the list
	if m.selected == nil {
		m.servers, serversCmd = m.servers.Update(msg)
	} else {
		m.addons, addonsCmd = m.addons.Update(msg)
	}

	cmds = append(
		cmds,
		cmd,
		serversCmd,
		addonsCmd,
	)

	return m, tea.Batch(cmds...)
}

type installEventMsg manager.Event

type installDoneMsg struct {
	err error
}

func (m Model) installMods(mods []game.Mod, ch chan<- tea.Msg) tea.Cmd {
	return func() tea.Msg {
		err := m.manager.Install(m.ctx, mods, func(e manager.Event) {
			ch <- installEventMsg(e)
		})
		ch <- installDoneMsg{err: err}
		return nil
	}
}

func waitInstall(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m Model) pendingMods() []game.Mod {
	var mods []game.Mod
	for _, it := range m.addons.Items() {
		if a := it.(addon); a.state != addonDone {
			mods = append(mods, game.Mod{ModID: a.id, Name: a.name, Version: a.target})
		}
	}
	return mods
}

func (m *Model) updateAddon(e manager.Event) tea.Cmd {
	items := slices.Clone(m.addons.Items())
	i := slices.IndexFunc(items, func(it list.Item) bool {
		return it.(addon).id == e.ModID
	})
	if i < 0 {
		return nil
	}
	a := items[i].(addon)
	items = slices.Delete(items, i, i+1)
	switch {
	case e.Err != nil:
		a.state, a.err = addonFailed, e.Err
		items = append(items, a)
	case e.Done:
		a.state = addonDone
		items = append(items, a)
	default:
		a.state, a.bar = addonDownloading, m.bar
		if e.Total > 0 {
			a.percent = float64(e.Downloaded) / float64(e.Total)
		}
		items = slices.Insert(items, 0, list.Item(a))
	}
	cmd := m.addons.SetItems(items)
	m.setTitle()
	return cmd
}

func (m *Model) setTitle() {
	items := m.addons.Items()
	done := 0
	for _, it := range items {
		if it.(addon).state == addonDone {
			done++
		}
	}
	m.addons.Title = fmt.Sprintf("Syncing %s (%d/%d)", m.selected.Name, done, len(items))
}

func (m *Model) finishInstall(err error) tea.Cmd {
	if err == nil {
		err = errors.New("not downloaded")
	}
	items := slices.Clone(m.addons.Items())
	for i, it := range items {
		if a := it.(addon); a.state == addonPending || a.state == addonDownloading {
			a.state, a.err = addonFailed, err
			items[i] = a
		}
	}
	return m.addons.SetItems(items)
}

type backMsg struct{}

type serverSelectedMsg struct {
	Server *game.Server
}

func selectServer(s *server) tea.Cmd {
	return func() tea.Msg {
		return serverSelectedMsg{
			Server: &s.Server,
		}
	}
}

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true

	//var s strings.Builder

	// unrecoverable state, print error and return
	if m.fatal != nil {
		v.SetContent(m.fatalScreen())
		return v
	}

	//s.WriteString(m.servers.View())

	content := m.servers.View()
	if m.selected != nil {
		content = m.addons.View()
		if len(m.addons.Items()) == 0 {
			w, h := m.sizes.GetContentSize()
			content = lipgloss.JoinVertical(lipgloss.Left,
				m.addons.Styles.TitleBar.Render(m.addons.Styles.Title.Render(m.addons.Title)),
				lipgloss.Place(w, h-2, lipgloss.Center, lipgloss.Center, "Nothing to sync"),
			)
		}
	}

	v.SetContent(styles.ScreenBorder.Render(content))
	return v
}

type favServersMsg struct {
	Servers []game.Server
}

func (m Model) fetchFavServers() func() tea.Msg {
	return func() tea.Msg {
		if m.game == nil {
			return fatalMsg{
				Msg: "Game not initialized, what the hell?",
			}
		}
		s, err := m.game.GetFavoriteServers(m.ctx)
		if err != nil {
			return fatalMsg{
				Msg: fmt.Sprintf("Unable to fetch favorite servers: %s", err.Error()),
			}
		}

		return favServersMsg{
			Servers: s,
		}
	}
}

func (m Model) fatalScreen() string {
	if m.fatal == nil {
		return ""
	}

	w, h := m.sizes.GetContentSize()

	return styles.ScreenBorder.
		BorderForeground(colors.ErrorText.Dark).
		Render(lipgloss.Place(
			w, h,
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.NewStyle().Foreground(colors.ErrorText).Render(m.fatal.Error()),
		))
}

type initMsg struct {
	Game     *game.Client
	Workshop *workshop.Client
	Manager  *manager.Manager
}

func (m Model) initDeps() tea.Msg {
	gd, err := config.GameDir()
	if err != nil {
		log.Error("failed to get game dir", "err", err)
		return fatalMsg{
			Msg: err.Error(),
		}
	}
	ad, err := config.AddonsDir()
	if err != nil {
		log.Error("failed to get addons dir", "err", err)
		return fatalMsg{
			Msg: err.Error(),
		}
	}

	ticket, release, err := steam.GetAuthSessionTicketDetached(gd)
	if err != nil {
		log.Error("failed to get steam session ticket", "err", err)
		return fatalMsg{
			Msg: "Could not get a Steam session ticket. Make sure Steam is running.",
		}
	}
	log.Debug("got steam session ticket", "ticket", ticket)

	a := auth.New(ticket)

	// Important: if we try to create an access token too fast Bohemia's backend fails to autehnticate/validate it.
	time.Sleep(150 * time.Millisecond)

	at, err := a.GetAccessToken(m.ctx)
	release()
	if err != nil {
		log.Error("failed to get access token", "err", err)
		return fatalMsg{
			Msg: "Could not get access token. Please try again later.",
		}
	}
	log.Debug("got access token", "token", at)

	ws := workshop.New(a)

	man, err := manager.New(ad, ws)
	if err != nil {
		log.Error("failed to initialize manager", "err", err)
		return fatalMsg{
			Msg: "Could not initialize addons manager.",
		}
	}

	return initMsg{
		Game:    game.New(a),
		Manager: man,
	}
}

type fatalMsg struct {
	Msg string
}

func (f fatalMsg) Error() string {
	return f.Msg
}

type sizes struct {
	screenWidth  int
	screenHeight int
}

func (s sizes) SetScreenSize(width int, height int) sizes {
	s.screenWidth = width
	s.screenHeight = height
	return s
}

func (s sizes) GetContentSize() (int, int) {
	return s.screenWidth, s.screenHeight
}
