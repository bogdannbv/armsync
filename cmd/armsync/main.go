package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/bogdannbv/armsync/internal/config"
	"github.com/bogdannbv/armsync/internal/steam"
	"github.com/bogdannbv/armsync/internal/tui"
	"github.com/urfave/cli/v3"
)

const (
	appName = "armsync"
	version = "0.1.0"
)

var debugLog *os.File

func main() {
	// Must run first: the process re-executes itself to obtain a Steam ticket.
	steam.RunTicketChild()

	cmd := &cli.Command{
		Name:    appName,
		Usage:   "Arma Reforger mod syncer",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "write debug logs to debug.log in the current directory",
			},
		},
		Before: setup,
		After:  teardown,
		Action: runTUI,
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "Choose the addons directory (and the game directory if it cannot be detected)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:      "addons",
						Aliases:   []string{"a"},
						Usage:     "addons `DIR`; skips the interactive picker",
						TakesFile: true,
					},
					&cli.StringFlag{
						Name:      "game",
						Aliases:   []string{"g"},
						Usage:     "Arma Reforger install `DIR`; skips auto-detection",
						TakesFile: true,
					},
				},
				Action: runConfig,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func setup(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Bool("debug") {
		f, err := os.OpenFile("debug.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return ctx, fmt.Errorf("could not open debug.log: %w", err)
		}
		debugLog = f
		log.SetOutput(f)
		log.SetTimeFormat(time.TimeOnly)
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetOutput(io.Discard)
	}

	if err := config.Load(); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func teardown(context.Context, *cli.Command) error {
	if debugLog != nil {
		return debugLog.Close()
	}
	return nil
}

func runTUI(ctx context.Context, _ *cli.Command) error {
	if !config.IsValid() {
		if err := configure(config.Get(), true); err != nil {
			return err
		}
	}

	if _, err := tea.NewProgram(tui.NewModel(ctx)).Run(); err != nil {
		return fmt.Errorf("could not run program: %w", err)
	}
	return nil
}

func runConfig(_ context.Context, cmd *cli.Command) error {
	cfg := config.Get()
	interactive := true

	if g := cmd.String("game"); g != "" {
		abs, err := filepath.Abs(g)
		if err != nil {
			return err
		}
		if err = config.ValidateGameDir(abs); err != nil {
			return err
		}
		cfg.GameDir = abs
	}
	if a := cmd.String("addons"); a != "" {
		abs, err := filepath.Abs(a)
		if err != nil {
			return err
		}
		cfg.AddonsDir = abs
		interactive = false
	}

	return configure(cfg, interactive)
}

// configure fills in the game dir (auto-detected when possible) and the addons
// dir (via the picker when askAddons is set), validates, and saves the config.
func configure(cfg config.Config, askAddons bool) error {
	var err error

	if cfg.GameDir, err = resolveGameDir(cfg.GameDir); err != nil {
		return err
	}

	if askAddons {
		if cfg.AddonsDir, err = pickAddonsDir(cfg.AddonsDir); err != nil {
			return err
		}
	}

	if err = config.SaveConfig(cfg); err != nil {
		return err
	}

	path, _ := config.Path()
	fmt.Printf("Game dir:   %s\nAddons dir: %s\nSaved to    %s\n", cfg.GameDir, cfg.AddonsDir, path)
	return nil
}

// resolveGameDir keeps a valid configured dir, otherwise auto-detects, and
// only falls back to the picker when detection fails.
func resolveGameDir(current string) (string, error) {
	if config.ValidateGameDir(current) == nil {
		return current, nil
	}
	if dir, err := config.FindGameDir(); err == nil {
		return dir, nil
	}

	start, ok := tui.FirstExistingDir(steamDirCandidates()...)
	if !ok {
		return "", errors.New("could not find a directory to start browsing from")
	}
	dir, err := tui.RunDirPicker(
		"Where is Arma Reforger installed?",
		"It could not be detected automatically. Navigate to the folder containing steam_appid.txt and press space.",
		start,
		config.ValidateGameDir,
	)
	if err != nil {
		return "", fmt.Errorf("game directory: %w", err)
	}
	return dir, nil
}

func pickAddonsDir(current string) (string, error) {
	hint := "Navigate to your Arma Reforger addons folder and press space to choose it."
	candidates := []string{current}
	if detected, err := config.FindAddonsDir(); err == nil {
		candidates = append(candidates, detected)
		if current == "" || current == detected {
			hint = "Detected the default addons folder. Press space to use it, or navigate elsewhere."
		}
	}
	candidates = append(candidates, config.FallbackStartDirs()...)

	start, ok := tui.FirstExistingDir(candidates...)
	if !ok {
		return "", errors.New("could not find a directory to start browsing from")
	}
	dir, err := tui.RunDirPicker("Choose the addons directory", hint, start, config.ValidateAddonsDir)
	if err != nil {
		return "", fmt.Errorf("addons directory: %w", err)
	}
	return dir, nil
}

func steamDirCandidates() []string {
	var c []string
	if libs, err := steam.Libraries(); err == nil {
		for _, lib := range libs {
			c = append(c, filepath.Join(lib, "steamapps", "common"))
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		c = append(c, home)
	}
	return c
}
