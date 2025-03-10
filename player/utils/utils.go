//go:build !windows

package playerutils

import (
	"log/slog"
	"os"
)

func KillProcess(p *os.Process, l *slog.Logger) error {
	if p == nil {
		return nil
	}

	pid := p.Pid
	l.Debug("killing process", "pid", pid)

	err := p.Kill()
	if err != nil {
		return err
	}

	l.Debug("killed process group", "pgid", pid)
	return nil
}
