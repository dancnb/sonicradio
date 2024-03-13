package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
	"github.com/dancnb/sonicradio/ui"
)

func main() {
	run()
}

func run() {
	cfg, _ := config.Load()

	var logW io.Writer
	if cfg.Debug && cfg.LogPath != "" {
		logFile := fmt.Sprintf("sonicradio-%d.log", time.Now().UnixMilli())
		lp := filepath.Join(cfg.LogPath, logFile)
		lp = "__debug.log" // dev only
		f, err := os.Create(lp)
		if err != nil {
			panic("could not create log file " + lp)
		}
		defer f.Close()
		logW = f
	} else {
		logW = io.Discard
	}
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(logW, opts)
	logger := slog.New(handler)
	log.SetFlags(log.Flags() &^ (log.Ldate))
	slog.SetDefault(logger)
	slog.Info("----------------------Starting----------------------")
	slog.Debug("loaded", "config", fmt.Sprintf("%#v", cfg))

	b := browser.NewApi(cfg)
	p := player.NewMPV()

	if _, err := ui.NewProgram(&cfg, b, p).Run(); err != nil {
		slog.Info(fmt.Sprintf("Error running program: %s", err.Error()))
		os.Exit(1)
	}
}
