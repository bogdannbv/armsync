//go:build windows

package steam

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bogdannbv/armsync/internal/utils"
	"golang.org/x/sys/windows/registry"
)

func FindSteamDir() (string, error) {
	dir, err := findSteamDirSimple()
	if err == nil {
		return dir, nil
	}

	return findSteamDirRegistry()
}

func findSteamDirRegistry() (string, error) {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Valve\Steam`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		// Try x86 variant
		key, err = registry.OpenKey(
			registry.LOCAL_MACHINE,
			`SOFTWARE\WOW6432Node\Valve\Steam`,
			registry.QUERY_VALUE,
		)
		if err != nil {
			return "", err
		}
	}
	defer key.Close()

	dir, _, err := key.GetStringValue("InstallPath")
	if err != nil {
		return "", err
	}

	return utils.ValidatePathIsDir(dir)
}

func findSteamDirSimple() (string, error) {
	dir := os.Getenv("programfiles(x86)")
	if dir == "" {
		return "", fmt.Errorf("could not find programfiles(x86)")
	}
	dir = filepath.Join(dir, "Steam")
	return utils.ValidatePathIsDir(dir)
}
