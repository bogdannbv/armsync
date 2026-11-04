//go:build !windows

package steam

import (
	"os"
	"path/filepath"

	"github.com/bogdannbv/armsync/internal/utils"
)

func FindSteamDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, ".local", "share", "Steam")
	return utils.ValidatePathIsDir(dir)
}
