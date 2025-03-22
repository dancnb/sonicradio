package playerutils

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

func KillProcess(p *os.Process, l *slog.Logger) error {
	if p == nil {
		return nil
	}

	pid := p.Pid
	l.Info("killing process", "pid", pid)

	cmd := exec.Command("taskkill", []string{"/F", "/T", "/PID", fmt.Sprintf("%d", pid)}...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		l.Error("taskkill cmd error", "error", cmd.Err.Error())
		return cmd.Err
	}
	err := cmd.Run()
	if err != nil {
		l.Error("taskkill cmd run", "error", err)
		return err
	}

	l.Info("killed process group", "pgid", pid)
	return nil
}
