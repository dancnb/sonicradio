//go:build !windows

package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"syscall"
)

// findProcess
//
// On Unix systems, os.FindProcess always succeeds and returns a Process for the given pid, regardless of whether the process exists.
// To test whether the process actually exists, see whether p.Signal(syscall.Signal(0)) reports an error.
func findProcess(pid int) bool {
	log := slog.With("method", "config.findProcess")
	p, _ := os.FindProcess(pid)
	if err := p.Signal(syscall.Signal(0)); !errors.Is(err, os.ErrProcessDone) {
		log.Info(fmt.Sprintf("existing pid %d running: %v", pid, err))
		return true
	} else {
		log.Info(fmt.Sprintf("existing pid=%v is not running anymore", pid))
		return false
	}
}
