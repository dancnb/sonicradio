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
		slog.Debug("killing process", "pid", m.cmd.Process.Pid)

		// err := m.cmd.Process.Kill()
		// if err != nil {
		// 	return err
		// }

		// pid := m.cmd.Process.Pid
		pid, err := syscall.Getpgid(m.cmd.Process.Pid)
		if err != nil {
			slog.Error("error getting process group", "pid", m.cmd.Process.Pid)
			return err
		}
		// err = syscall.Kill(-m.cmd.Process.Pid, syscall.SIGCHLD)
		// if err != nil {
		// 	slog.Error("error killing process group", "pgid", pid)
		// 	return err
		// }
		err = syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			slog.Error("error killing process children", "pgid", pid)
			return err
		}

		slog.Debug("killed process group", "pgid", pid)
	}

	return nil
}
