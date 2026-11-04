package tui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

type Colors struct {
	ErrorText compat.AdaptiveColor
}

var colors = &Colors{
	ErrorText: compat.AdaptiveColor{
		Light: lipgloss.Color("#FD4B4B"),
		Dark:  lipgloss.Color("#FD4B4B"),
	},
}
