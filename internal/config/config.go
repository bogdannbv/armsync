package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func Load() error {
	f, err := initAndGetConfigFile()
	if err != nil {
		return fmt.Errorf("could not initialize config file: %w", err)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return SaveConfig(Config{})
		}
		return fmt.Errorf("could not read config file: %w", err)
	}

	loaded.Store(true)

	return nil
}

func IsLoaded() bool {
	return loaded.Load()
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

	if err := Validate(config); err != nil {
		return false
	}

	return true
}

func Validate(cfg Config) error {
	if cfg.GameDir == "" || cfg.AddonsDir == "" {
		return errors.New("invalid dirs")
	}

	if err := steam.ValidateSteamAppIDPresence(cfg.GameDir); err != nil {
		return fmt.Errorf("unable to check that %s contains steam_appid.txt, please try again", cfg.GameDir)
	}

	if _, err := utils.ValidatePathIsDir(cfg.AddonsDir); err != nil {
		return err
	}

	return nil
}

func SaveConfig(cfg Config) error {
	if err := Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	f, err := initAndGetConfigFile()
	if err != nil {
		return fmt.Errorf("could not initialize config file: %w", err)
	}
	defer f.Close()

	if off, err := f.Seek(0, io.SeekStart); err != nil || off != 0 {
		return fmt.Errorf("could not seek to beginning of config file: %w", err)
	}
	if err = f.Truncate(0); err != nil {
		return fmt.Errorf("could not truncate config file: %w", err)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode config file: %w", err)

	}
	if _, err = f.Write(b); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	config = cfg
	loaded.Store(true)

	return nil
}

func initAndGetConfigFile() (*os.File, error) {
	path, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user config dir: %w", err)
	}
	path = filepath.Join(path, "armsync")
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("could not create config directory %s: %w", path, err)
	}
	path = filepath.Join(path, "config.json")

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return f, nil
}
