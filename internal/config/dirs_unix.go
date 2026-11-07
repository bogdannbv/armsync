//go:build !windows

package config

import (
	"os"
	"path/filepath"

	"github.com/bogdannbv/armsync/internal/steam"
	"github.com/bogdannbv/armsync/internal/utils"
	"github.com/bogdannbv/armsync/pkg/reforger/constants"
)

// FindAddonsDir returns the addons directory inside the game's Proton prefix.
func FindAddonsDir() (string, error) {
	lib, err := steam.FindAppLibrary(constants.AppID)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(
		lib,
		"steamapps",
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

// FallbackStartDirs lists directories, in order of preference, where the
// directory picker should open when the addons directory cannot be detected.
func FallbackStartDirs() []string {
	var dirs []string
	if lib, err := steam.FindAppLibrary(constants.AppID); err == nil {
		dirs = append(dirs, filepath.Join(lib, "steamapps", "compatdata", constants.AppID, "pfx", "drive_c"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, home)
	}
	return dirs
}
