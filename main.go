package main

import (
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Println(err.Error())
	}
	if cfg.IsRunning {
		fmt.Println("application is already running")
		os.Exit(1)
	}
	cfg.IsRunning = true
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			cfg.IsRunning = false
			err = cfg.Save()
			if err != nil {
				fmt.Printf("error saving config: %v\n", err)
			}
		}
	}()

	var logW io.Writer
	if cfg.Debug {
		logFile := fmt.Sprintf("sonicradio-%d.log", time.Now().UnixMilli())
		lp := filepath.Join(cfg.LogPath, logFile)
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
	slog.Debug("loaded", "config", cfg.String())

	b, err := browser.NewApi(ctx, cfg)
	if err != nil {
		panic(err)
	}
	p, err := player.NewPlayer(ctx, cfg)
	if err != nil {
		panic(err)
	}
	m := ui.NewModel(ctx, cfg, b, p)
	defer func() {
		m.Quit()
	}()

	err = cfg.Save()
	if err != nil {
		fmt.Printf("error saving config: %v\n", err)
	}

	if _, err := m.Progr.Run(); err != nil {
		slog.Info(fmt.Sprintf("Error running program: %s", err.Error()))
		cfg.IsRunning = false
		err = cfg.Save()
		if err != nil {
			fmt.Printf("error saving config: %v\n", err)
		}
		os.Exit(1)
	}
}
