package config

import "github.com/bogdannbv/armsync/internal/steam"

// FindGameDir returns the detected Arma Reforger installation directory.
func FindGameDir() (string, error) {
	return steam.FindGameDir()
}
