package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

const pidFileName = "sonicradio.pid"

var ErrInstanceRunning = errors.New("application is already running")

func CheckPidFile() (*os.File, error) {
	log := slog.With("method", "config.CheckPidFile")
	pid := os.Getpid()
	log.Info(fmt.Sprintf("current pid=%v", pid))

	cfgDir, err := getOrCreateConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get or create config dir error: %v", err)
	}

	fp := filepath.Join(cfgDir, pidFileName)
	_, err = os.Stat(fp)

	if err == nil {
		log.Info("found existing pid file, checking pid")
		b, err := os.ReadFile(fp)
		if err != nil {
			return nil, fmt.Errorf("read existing pid file %q error: %v", fp, err)
		}
		log.Info(fmt.Sprintf("found existing pid=%s", b))
		exPid, err := strconv.Atoi(string(b))
		if err != nil {
			return nil, fmt.Errorf("parse existing pid file %q, content: %q, error: %v", fp, b, err)
		}

		isRunning := findProcess(exPid)
		if isRunning {
			return nil, ErrInstanceRunning
		}
		return createPidFile(fp, pid)

	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error stat pid file: %v", err)
	}

	log.Info("no existing pid file")
	return createPidFile(fp, pid)
}

func createPidFile(fp string, pid int) (*os.File, error) {
	log := slog.With("method", "config.createPidFile")
	log.Info("pid file not found, creating", "pid", pid)
	f, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprint(f, pid); err != nil {
		return nil, err
	}
	return f, nil
}
