package tui

import (
	"fmt"

	"github.com/bogdannbv/armsync/pkg/reforger/game"
)

type server struct {
	game.Server
}

func (s server) Title() string {
	return s.Name
}

func (s server) Description() string {
	return fmt.Sprintf(
		"Players: %d/%d | Queue: %d/%d | %s | %s",
		s.PlayerCount, s.PlayerCountLimit,
		s.JoinQueue.Size, s.JoinQueue.MaxSize,
		s.GameVersion,
		s.HostAddress,
	)
}

func (s server) FilterValue() string {
	return s.Name
}
