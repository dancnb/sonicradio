package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
	"github.com/dancnb/sonicradio/ui"
)

var (
	version = "0.8.5"

	cpuProfile  = flag.String("cpuprofile", "", "-cpuprofile=<filename>")
	memProfile  = flag.String("memprofile", "", "-memprofile=<filename>")
	httpProfile = flag.Bool("httpprofile", false, "-httpprofile")
)

func main() {
	flag.Parse()

	if cpuProfile != nil && *cpuProfile != "" {
		f, err := os.Create(fmt.Sprintf("%s_%d", *cpuProfile, time.Now().Unix()))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		run()

	} else if memProfile != nil && *memProfile != "" {
		f, err := os.Create(fmt.Sprintf("%s_%d", *memProfile, time.Now().Unix()))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		runtime.MemProfileRate = 1

		run()

		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}

	} else if httpProfile != nil && *httpProfile {
		go func() {
			log.Println(http.ListenAndServe(":6060", nil))
		}()

		run()

	} else {
		run()
	}
}

func run() {
	logWC := createLogger()
	defer func() {
		_ = logWC.Close()
	}()

	pidFile, err := config.CheckPidFile()
	if err != nil {
		fmt.Printf("check running instance: %v\n", err)
		_ = logWC.Close()
		os.Exit(1)
	}
	defer func() {
		if err := os.Remove(pidFile.Name()); err != nil {
			slog.Error(fmt.Sprintf("error removing pid file: %v", err))
		}
	}()

	slog.Info("----------------------Starting----------------------")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load(version)
	if err != nil {
		slog.Info("load config", "error", err.Error())
	}
	if cfg == nil {
		panic("could not get config")
	}

	slog.Info("loaded", "config", cfg.String())

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

	if _, err := m.Progr.Run(); err != nil {
		slog.Info(fmt.Sprintf("Error running program: %s", err.Error()))
	}
}

type nopWriterCloser struct {
	io.Writer
}

func (n nopWriterCloser) Close() error { return nil }

func createLogger() io.WriteCloser {
	var logW io.WriteCloser
	if config.Debug() {
		logFilePath := fmt.Sprintf("sonicradio-%d.log", time.Now().UnixMilli())
		logFilePath = filepath.Join(os.TempDir(), logFilePath)
		logFile, err := os.Create(logFilePath)
		if err != nil {
			panic("could not create log file " + logFilePath)
		}
		logW = logFile
	} else {
		logW = nopWriterCloser{io.Discard}
	}
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(logW, opts)
	logger := slog.New(handler)
	log.SetFlags(log.Flags() &^ (log.Ldate))
	slog.SetDefault(logger)

	return logW
}
