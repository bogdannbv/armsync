//go:build windows

package manager

import "path/filepath"

// On windows it's safe to just join them
func getPackageFilePath(dir, fn string) string {
	return filepath.Join(dir, fn)
}
