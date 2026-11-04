package tui

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	ScreenBorder lipgloss.Style
}

var styles = Styles{
	ScreenBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true),
}
