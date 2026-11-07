package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ErrPickerCancelled is returned by RunDirPicker when the user quits without choosing.
var ErrPickerCancelled = errors.New("directory selection cancelled")

type pickerKeyMap struct {
	Choose   key.Binding
	TypePath key.Binding
	Open     key.Binding
	Up       key.Binding
	Cancel   key.Binding
	Confirm  key.Binding
	Yes      key.Binding
	No       key.Binding
}

var pickerKeys = pickerKeyMap{
	Choose:   key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "choose this folder")),
	TypePath: key.NewBinding(key.WithKeys("/", "e"), key.WithHelp("/", "type a path")),
	Open:     key.NewBinding(key.WithKeys("enter", "right", "l"), key.WithHelp("enter/→", "open")),
	Up:       key.NewBinding(key.WithKeys("backspace", "left", "h"), key.WithHelp("←/backspace", "up")),
	Cancel:   key.NewBinding(key.WithKeys("esc", "ctrl+c", "q"), key.WithHelp("esc", "cancel")),
	Confirm:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "go")),
	Yes:      key.NewBinding(key.WithKeys("y", "Y", "enter"), key.WithHelp("y/enter", "yes")),
	No:       key.NewBinding(key.WithKeys("n", "N", "esc", "space"), key.WithHelp("n/esc", "go back")),
}

var (
	pickerTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	pickerHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	pickerPathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	pickerErrStyle   = lipgloss.NewStyle().Foreground(colors.ErrorText)
	pickerHelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dialogStyle      = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Padding(1, 2)
	dialogTitleStyle = lipgloss.NewStyle().Bold(true)
	dialogPathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
)

// DirPicker is a full-screen directory chooser built on bubbles' filepicker.
type DirPicker struct {
	title    string
	hint     string
	validate func(string) error

	fp     filepicker.Model
	input  textinput.Model
	typing bool

	width, height int
	errMsg        string

	// confirming is set after the user picked a folder and must confirm it.
	confirming bool
	pending    string

	chosen    string
	cancelled bool
}

// NewDirPicker creates a picker that starts in startDir. validate, if not nil,
// is run against the chosen directory and its error shown to the user.
func NewDirPicker(title, hint, startDir string, validate func(string) error) DirPicker {
	fp := filepicker.New()
	fp.CurrentDirectory = startDir
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = true // ~/.local on Linux must be reachable
	fp.ShowPermissions = false
	fp.ShowSize = false
	fp.AutoHeight = false
	fp.KeyMap.Open = pickerKeys.Open
	fp.KeyMap.Back = pickerKeys.Up
	// Selection is handled by this model via Choose; disable the picker's own.
	fp.KeyMap.Select = key.NewBinding(key.WithDisabled())
	fp.Styles.EmptyDirectory = fp.Styles.EmptyDirectory.SetString("(empty folder)")

	in := textinput.New()
	in.Prompt = "Path: "
	in.Placeholder = "type or paste a directory path"

	return DirPicker{
		title:    title,
		hint:     hint,
		validate: validate,
		fp:       fp,
		input:    in,
	}
}

func (m DirPicker) Init() tea.Cmd {
	return m.fp.Init()
}

func (m DirPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.fp.SetHeight(max(1, msg.Height-m.chromeHeight()))
		m.input.SetWidth(max(10, msg.Width-len(m.input.Prompt)-2))
		// Fall through: the filepicker recomputes its visible window from the new height.
	case tea.KeyPressMsg:
		if m.confirming {
			return m.updateConfirming(msg)
		}
		if m.typing {
			return m.updateTyping(msg)
		}
		switch {
		case key.Matches(msg, pickerKeys.Cancel):
			m.cancelled = true
			return m, tea.Quit
		case key.Matches(msg, pickerKeys.Choose):
			return m.choose(m.fp.CurrentDirectory)
		case key.Matches(msg, pickerKeys.TypePath):
			m.typing = true
			m.errMsg = ""
			m.input.SetValue(m.fp.CurrentDirectory + string(filepath.Separator))
			m.input.CursorEnd()
			return m, m.input.Focus()
		}
		m.errMsg = ""
	}

	var cmd tea.Cmd
	m.fp, cmd = m.fp.Update(msg)
	return m, cmd
}

func (m DirPicker) updateConfirming(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, pickerKeys.Yes):
		m.chosen = m.pending
		return m, tea.Quit
	case key.Matches(msg, pickerKeys.No):
		m.confirming = false
		m.pending = ""
	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "q"))):
		m.cancelled = true
		return m, tea.Quit
	}
	return m, nil
}

func (m DirPicker) updateTyping(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, pickerKeys.Cancel):
		m.typing = false
		m.input.Blur()
		m.errMsg = ""
		return m, nil
	case key.Matches(msg, pickerKeys.Confirm):
		p := expandPath(strings.TrimSpace(m.input.Value()))
		if p == "" {
			return m, nil
		}
		if err := checkReadableDir(p); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.typing = false
		m.input.Blur()
		m.errMsg = ""
		m.fp.CurrentDirectory = p
		return m, m.fp.Init()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m DirPicker) choose(dir string) (tea.Model, tea.Cmd) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		m.errMsg = err.Error()
		return m, nil
	}
	if m.validate != nil {
		if err = m.validate(abs); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
	}
	m.pending = abs
	m.confirming = true
	return m, nil
}

func (m DirPicker) chromeHeight() int {
	// title, hint, blank, path line, blank, [list], blank, error, help
	return 9
}

func (m DirPicker) View() tea.View {
	var v tea.View
	v.AltScreen = true

	var b strings.Builder
	b.WriteString(pickerTitleStyle.Render(m.title) + "\n")
	b.WriteString(pickerHintStyle.Render(m.hint) + "\n\n")

	if m.typing {
		b.WriteString(m.input.View() + "\n\n")
		v.Cursor = m.input.Cursor()
	} else {
		b.WriteString(pickerPathStyle.Render("📁 "+m.fp.CurrentDirectory) + "\n\n")
	}

	b.WriteString(m.fp.View())
	b.WriteString("\n")

	if m.errMsg != "" {
		b.WriteString(pickerErrStyle.Render(m.errMsg))
	}
	b.WriteString("\n")

	var help []key.Binding
	if m.typing {
		help = []key.Binding{pickerKeys.Confirm, key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back to list"))}
	} else {
		help = []key.Binding{pickerKeys.Open, pickerKeys.Up, pickerKeys.Choose, pickerKeys.TypePath, pickerKeys.Cancel}
	}
	b.WriteString(pickerHelpStyle.Render(renderHelp(help)))

	content := b.String()
	if m.confirming {
		content = m.overlayDialog(content)
	}
	v.SetContent(content)
	return v
}

// overlayDialog draws the confirmation dialog centred on top of base.
func (m DirPicker) overlayDialog(base string) string {
	w, h := max(m.width, 20), max(m.height, 5)

	inner := min(max(40, lipgloss.Width(m.pending)+2), w-8)
	body := lipgloss.JoinVertical(lipgloss.Left,
		dialogTitleStyle.Render("Use this folder?"),
		"",
		dialogPathStyle.Width(inner).Render(m.pending),
		"",
		pickerHelpStyle.Render(renderHelp([]key.Binding{pickerKeys.Yes, pickerKeys.No})),
	)
	dialog := dialogStyle.Width(inner).Render(body)

	// Pad the base to the full screen so the compositor's bounds match it.
	base = lipgloss.NewStyle().Width(w).Height(h).MaxHeight(h).Render(base)
	x := max(0, (w-lipgloss.Width(dialog))/2)
	y := max(0, (h-lipgloss.Height(dialog))/2)

	return lipgloss.NewCompositor(
		lipgloss.NewLayer(base),
		lipgloss.NewLayer(dialog).X(x).Y(y).Z(1),
	).Render()
}

func renderHelp(bindings []key.Binding) string {
	parts := make([]string, 0, len(bindings))
	for _, kb := range bindings {
		h := kb.Help()
		parts = append(parts, fmt.Sprintf("%s %s", h.Key, h.Desc))
	}
	return strings.Join(parts, " • ")
}

// Result returns the chosen directory, or ErrPickerCancelled.
func (m DirPicker) Result() (string, error) {
	if m.cancelled || m.chosen == "" {
		return "", ErrPickerCancelled
	}
	return m.chosen, nil
}

// RunDirPicker shows the picker and blocks until the user chooses or cancels.
func RunDirPicker(title, hint, startDir string, validate func(string) error) (string, error) {
	final, err := tea.NewProgram(NewDirPicker(title, hint, startDir, validate)).Run()
	if err != nil {
		return "", err
	}
	m, ok := final.(DirPicker)
	if !ok {
		return "", errors.New("unexpected picker model")
	}
	return m.Result()
}

// FirstExistingDir returns the first candidate that is a readable directory.
func FirstExistingDir(candidates ...string) (string, bool) {
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if checkReadableDir(c) == nil {
			return c, true
		}
	}
	return "", false
}

func checkReadableDir(p string) error {
	fi, err := os.Stat(p)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", p)
	}
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	return f.Close()
}

func expandPath(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		if home, err := os.UserHomeDir(); err == nil {
			p = home + p[1:]
		}
	}
	return filepath.Clean(p)
}
