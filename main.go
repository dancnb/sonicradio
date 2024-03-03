package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
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
	logger := slog.New(slog.NewJSONHandler(logW, nil))
	slog.SetDefault(logger)
	slog.Info("----Starting----")

	b := browser.NewApi(cfg)

	if _, err := ui.NewProgram(cfg, b).Run(); err != nil {
		slog.Info(fmt.Sprintf("Error running program: %s", err.Error()))
		os.Exit(1)
	}
}
