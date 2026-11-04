package tui

import (
	"fmt"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
)

var checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")

type addonState int

const (
	addonPending addonState = iota
	addonDownloading
	addonDone
	addonFailed
)

type addon struct {
	id      string
	name    string
	version string
	target  string
	state   addonState
	percent float64
	bar     progress.Model
	err     error
}

func (a addon) Title() string {
	return a.name
}

func (a addon) Description() string {
	switch a.state {
	case addonDownloading:
		return a.bar.ViewAs(a.percent)
	case addonDone:
		return checkMark + " " + a.versions()
	case addonFailed:
		return "✗ " + a.err.Error()
	}
	return a.versions()
}

func (a addon) versions() string {
	if a.version == "" {
		return a.target
	}
	return fmt.Sprintf("%s -> %s", a.version, a.target)
}

func (a addon) FilterValue() string {
	return a.name
}
