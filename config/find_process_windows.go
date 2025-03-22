package config

import (
	"fmt"
	"log/slog"
	"os"
)

// findProcess
//
// On Windows, it returns a non-nil os.Porcess if it is runnning, otherwise a nil os.Process and an error
func findProcess(pid int) bool {
	log := slog.With("caller", "config.findProcess")
	if p, err := os.FindProcess(pid); err == nil && p != nil {
		log.Info(fmt.Sprintf("existing pid %d running: %v", pid, err))
		return true
	} else {
		log.Error(fmt.Sprintf("error finding existing pid %d : %v", pid, err))
		return false
	}
}
