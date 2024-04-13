package player

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"time"
)

const (
	baseCmd         = "mpv"
	errOut          = "Failed to"
	titleMsg        = "icy-title:"
	startWaitMillis = 500
)

var baseArgs = []string{"--no-video", "--quiet"}

func NewMPV() Player {
	return &Mpv{}
}

type Mpv struct {
	cmd    *exec.Cmd
	cmdOut bytes.Buffer
}

func (m *Mpv) Play(url string) (string, error) {
	slog.Info("playing url=" + url)

	if err := m.Stop(); err != nil {
		return "", err
	}

	args := slices.Clone(baseArgs)
	args = append(args, url)
	cmd := exec.Command(baseCmd, args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		slog.Error("mpv cmd error", "error", cmd.Err.Error())
		return "", cmd.Err
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	m.cmdOut.Reset()
	cmd.Stdout = &m.cmdOut

	err := cmd.Start()
	if err != nil {
		slog.Error("mpv cmd start", "error", err)
		return "", err
	}
	time.Sleep(startWaitMillis * time.Millisecond)
	b, err := io.ReadAll(&m.cmdOut)
	if err != nil {
		return "", err
	}
	output := string(b)
	slog.Debug("mpv cmd", "output", output)
	errIx := strings.Index(output, errOut)
	if errIx >= 0 {
		errMsg := output[errIx:]
		nlIx := strings.Index(errMsg, "\n")
		if nlIx >= 0 {
			errMsg = errMsg[:nlIx]
		}
		return "", errors.New(errMsg)
	}
	title := ""
	titleIx := strings.Index(output, titleMsg)
	if titleIx >= 0 {
		title = output[titleIx+10:]
		nlIx := strings.Index(title, "\n")
		if nlIx >= 0 {
			title = title[:nlIx]
		}
	}
	title = strings.TrimSpace(title)

	m.cmd = cmd
	slog.Debug("mpv cmd started", "pid", m.cmd.Process.Pid)
	return title, nil
}

func (m *Mpv) Stop() error {
	if m.cmd == nil {
		slog.Debug("no current station playing")
		return nil
	}
	// err := m.cmd.Wait()
	// if err != nil {
	// 	return err
	// }

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
		m.cmd = nil
	}

	return nil
}
