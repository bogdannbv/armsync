//go:build windows

package config

import (
	"os"
	"path/filepath"

	"github.com/bogdannbv/armsync/internal/utils"
)

func FindAddonsDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "Documents", "My Games", "ArmaReforger", "addons")
	return utils.ValidatePathIsDir(dir)
}

// FallbackStartDirs lists directories, in order of preference, where the
// directory picker should open when the addons directory cannot be detected.
func FallbackStartDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, "Documents", "My Games"),
		filepath.Join(home, "Documents"),
		home,
	}
}
