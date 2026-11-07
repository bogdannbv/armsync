package steam

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/bogdannbv/armsync/internal/utils"
	"github.com/bogdannbv/armsync/pkg/reforger/constants"
)

const defaultInstallDir = "Arma Reforger"

var (
	// Matches `"path"  "..."` lines in libraryfolders.vdf.
	libraryPathRe = regexp.MustCompile(`(?m)^\s*"path"\s+"((?:[^"\\]|\\.)*)"`)
	// Matches `"installdir"  "..."` in an appmanifest_<id>.acf file.
	installDirRe = regexp.MustCompile(`(?m)^\s*"installdir"\s+"((?:[^"\\]|\\.)*)"`)
)

// Libraries returns the roots of every Steam library known to the local
// Steam installation. The main Steam directory is always first.
func Libraries() ([]string, error) {
	root, err := FindSteamDir()
	if err != nil {
		return nil, err
	}
	libs := []string{filepath.Clean(root)}

	b, err := os.ReadFile(filepath.Join(root, "steamapps", "libraryfolders.vdf"))
	if err != nil {
		// No library file: the main directory is the only library.
		return libs, nil
	}
	for _, m := range libraryPathRe.FindAllStringSubmatch(string(b), -1) {
		p := filepath.Clean(unescapeVDF(m[1]))
		if !slices.Contains(libs, p) {
			libs = append(libs, p)
		}
	}
	return libs, nil
}

// FindAppLibrary returns the Steam library root in which the given app is installed.
func FindAppLibrary(appID string) (string, error) {
	libs, err := Libraries()
	if err != nil {
		return "", err
	}
	for _, lib := range libs {
		if _, err := os.Stat(appManifestPath(lib, appID)); err == nil {
			return lib, nil
		}
	}
	return "", fmt.Errorf("steam app %s is not installed in any steam library", appID)
}

// FindGameDir locates the Arma Reforger installation directory.
func FindGameDir() (string, error) {
	lib, err := FindAppLibrary(constants.AppID)
	if err != nil {
		return "", err
	}

	installDir := defaultInstallDir
	if b, err := os.ReadFile(appManifestPath(lib, constants.AppID)); err == nil {
		if m := installDirRe.FindStringSubmatch(string(b)); m != nil && m[1] != "" {
			installDir = unescapeVDF(m[1])
		}
	}

	dir, err := utils.ValidatePathIsDir(filepath.Join(lib, "steamapps", "common", installDir))
	if err != nil {
		return "", err
	}
	if err = ValidateSteamAppIDPresence(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func appManifestPath(lib, appID string) string {
	return filepath.Join(lib, "steamapps", "appmanifest_"+appID+".acf")
}

func unescapeVDF(s string) string {
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = strings.ReplaceAll(s, `\"`, `"`)
	return s
}
