package player

import (
	"errors"
	"log/slog"
	"os/exec"
	"slices"
	"syscall"
)

const (
	baseCmd = "mpv"
)

var baseArgs = []string{"--no-video", "--really-quiet"}

func NewMPV() Player {
	return &Mpv{}
}

type Mpv struct {
	cmd *exec.Cmd
}

func (m *Mpv) Play(url string) error {
	slog.Info("playing url=" + url)

	if err := m.Stop(); err != nil {
		return err
	}

	args := slices.Clone(baseArgs)
	args = append(args, url)
	cmd := exec.Command(baseCmd, args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		slog.Error("mpv cmd error", "error", cmd.Err.Error())
		return cmd.Err
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	if err != nil {
		slog.Error("mpv cmd start", "error", err)
		return err
	}

	m.cmd = cmd
	slog.Debug("mpv cmd started", "pid", m.cmd.Process.Pid)
	return nil
}

func (m *Mpv) Stop() error {
	if m.cmd == nil {
		slog.Debug("no current station playing")
		return nil
	}

	if m.cmd.Process != nil {
		// m.cmd.Process.Kill()
		slog.Info("killing process", "pid", m.cmd.Process.Pid)
		pid := m.cmd.Process.Pid
		err := syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			slog.Error("error getting process group", "pid", m.cmd.Process.Pid)
			return err
		}
		slog.Debug("killed process", "pid", pid)
	}

	return nil
}
