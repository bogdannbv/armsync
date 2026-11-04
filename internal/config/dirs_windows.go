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
