//go:build windows

package steam

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bogdannbv/armsync/internal/utils"
	"golang.org/x/sys/windows/registry"
)

func FindSteamDir() (string, error) {
	var errs []error
	for _, find := range []func() (string, error){findSteamDirUserRegistry, findSteamDirMachineRegistry, findSteamDirSimple} {
		dir, err := find()
		if err == nil {
			return dir, nil
		}
		errs = append(errs, err)
	}
	return "", errors.Join(errs...)
}

// findSteamDirUserRegistry reads the per-user Steam path, which is the most reliable source.
func findSteamDirUserRegistry() (string, error) {
	return readRegistryDir(registry.CURRENT_USER, `Software\Valve\Steam`, "SteamPath")
}

func findSteamDirMachineRegistry() (string, error) {
	dir, err := readRegistryDir(registry.LOCAL_MACHINE, `SOFTWARE\Valve\Steam`, "InstallPath")
	if err == nil {
		return dir, nil
	}
	// Try x86 variant
	return readRegistryDir(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, "InstallPath")
}

func readRegistryDir(root registry.Key, path, name string) (string, error) {
	key, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	dir, _, err := key.GetStringValue(name)
	if err != nil {
		return "", err
	}
	return utils.ValidatePathIsDir(filepath.Clean(dir))
}

func findSteamDirSimple() (string, error) {
	dir := os.Getenv("programfiles(x86)")
	if dir == "" {
		return "", fmt.Errorf("could not find programfiles(x86)")
	}
	dir = filepath.Join(dir, "Steam")
	return utils.ValidatePathIsDir(dir)
}
