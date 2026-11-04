package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/bogdannbv/armsync/internal/config"
	"github.com/bogdannbv/armsync/internal/steam"
	"github.com/bogdannbv/armsync/internal/tui"
)

func main() {
	steam.RunTicketChild()
	debug := flag.Bool("debug", false, "write debug logs to debug.log")
	flag.Parse()

	if err := config.Load(); err != nil {
		log.Fatal("could not load config", "err", err)
	}
	if !config.IsValid() {
		log.Fatal("invalid config, run `armsync config` first")
	}

	if *debug {
		f, err := os.OpenFile("debug.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			log.Fatal("could not open debug.log", "err", err)
		}
		defer f.Close()
		log.SetOutput(f)
		log.SetTimeFormat(time.TimeOnly)
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetOutput(io.Discard)
	}

	if _, err := tea.NewProgram(tui.NewModel(context.Background())).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "could not run program:", err)
		os.Exit(1)
	}
}
