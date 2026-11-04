package steam

import (
	"path/filepath"

	"github.com/bogdannbv/armsync/internal/utils"
)

func FindSteamAppsDir() (string, error) {
	dir, err := FindSteamDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "steamapps")
	return utils.ValidatePathIsDir(dir)
}
