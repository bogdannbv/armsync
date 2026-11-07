package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/bogdannbv/armsync/internal/steam"
	"github.com/bogdannbv/armsync/internal/utils"
)

type Config struct {
	GameDir   string `json:"game_dir"`
	AddonsDir string `json:"addons_dir"`
}

var ErrNotLoaded = errors.New("config not loaded")

var loaded atomic.Bool
var config Config

// Load reads the config file. A missing or empty file is not an error: the
// config is then loaded with zero values and IsValid reports false.
func Load() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			config = Config{}
			loaded.Store(true)
			return nil
		}
		return fmt.Errorf("could not read config file %s: %w", path, err)
	}

	var cfg Config
	if len(b) > 0 {
		if err = json.Unmarshal(b, &cfg); err != nil {
			return fmt.Errorf("could not parse config file %s: %w", path, err)
		}
	}

	config = cfg
	loaded.Store(true)
	return nil
}

func IsLoaded() bool {
	return loaded.Load()
}

// Get returns a copy of the current config.
func Get() Config {
	return config
}

func GameDir() (string, error) {
	if !IsLoaded() {
		return "", ErrNotLoaded
	}
	return config.GameDir, nil
}

func AddonsDir() (string, error) {
	if !IsLoaded() {
		return "", ErrNotLoaded
	}
	return config.AddonsDir, nil
}

func IsValid() bool {
	if !IsLoaded() {
		return false
	}
	return Validate(config) == nil
}

func Validate(cfg Config) error {
	if err := ValidateGameDir(cfg.GameDir); err != nil {
		return err
	}
	return ValidateAddonsDir(cfg.AddonsDir)
}

func ValidateGameDir(dir string) error {
	if dir == "" {
		return errors.New("game directory is not set")
	}
	if err := steam.ValidateSteamAppIDPresence(dir); err != nil {
		return fmt.Errorf("%s does not look like the Arma Reforger directory: %w", dir, err)
	}
	return nil
}

func ValidateAddonsDir(dir string) error {
	if dir == "" {
		return errors.New("addons directory is not set")
	}
	_, err := utils.ValidatePathIsDir(dir)
	return err
}

func SaveConfig(cfg Config) error {
	if err := Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	path, err := configPath()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode config: %w", err)
	}

	// Write to a temp file and rename so a crash never leaves a truncated config.
	tmp, err := os.CreateTemp(filepath.Dir(path), "config-*.json.tmp")
	if err != nil {
		return fmt.Errorf("could not create temp config file: %w", err)
	}
	tmpPath := tmp.Name()
	_, werr := tmp.Write(b)
	cerr := tmp.Close()
	if err = errors.Join(werr, cerr); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("could not write config file: %w", err)
	}
	if err = os.Chmod(tmpPath, 0o644); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("could not set config file permissions: %w", err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("could not save config file: %w", err)
	}

	config = cfg
	loaded.Store(true)
	return nil
}

// Path returns the location of the config file.
func Path() (string, error) {
	return configPath()
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get user config dir: %w", err)
	}
	return filepath.Join(dir, "armsync", "config.json"), nil
}
