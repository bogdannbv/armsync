package config

import (
	"path/filepath"

	"github.com/bogdannbv/armsync/internal/steam"
	"github.com/bogdannbv/armsync/internal/utils"
)

func FindGameDir() (string, error) {
	dir, err := steam.FindSteamAppsDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "common", "Arma Reforger")
	return utils.ValidatePathIsDir(dir)
}
