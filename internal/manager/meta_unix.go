//go:build !windows

package manager

import (
	"fmt"
	"path"
)

// On linux it's a bit cursed?
// TODO allow some flexibility maybe?
func getPackageFilePath(dir, fn string) string {
	modDir := path.Base(dir)
	return fmt.Sprintf("$profile:../addons/%s/%s", modDir, fn)
}
