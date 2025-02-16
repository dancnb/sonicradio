package config

import (
	"testing"
)

func TestCheckProcess(t *testing.T) {
	pid, isRunning := CheckPidFile()
	t.Log(pid, isRunning)
}
