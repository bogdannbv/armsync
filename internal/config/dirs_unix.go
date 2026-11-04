//go:build !windows

package config

import (
	"path/filepath"

	"github.com/bogdannbv/armsync/internal/steam"
	"github.com/bogdannbv/armsync/internal/utils"
	"github.com/bogdannbv/armsync/pkg/reforger/constants"
)

func FindAddonsDir() (string, error) {
	dir, err := steam.FindSteamAppsDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(
		dir,
		"compatdata",
		constants.AppID,
		"pfx",
		"drive_c",
		"users",
		"steamuser",
		"Documents",
		"My Games",
		"ArmaReforger",
		"addons",
	)
	return utils.ValidatePathIsDir(dir)
}
