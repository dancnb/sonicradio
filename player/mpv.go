package player

import (
	"bytes"
	"errors"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"syscall"
)

const (
	baseCmd  = "mpv"
	errOut   = "Failed to"
	titleMsg = "icy-title:"
)

var baseArgs = []string{"--no-video", "--quiet"}

func NewMPV() Player {
	return &Mpv{}
}

type Mpv struct {
	cmd    *exec.Cmd
	output bytes.Buffer
}

func (mpv *Mpv) Play(url string) error {
	slog.Info("playing url=" + url)
	mpv.output.Reset()

	if err := mpv.Stop(); err != nil {
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
	cmd.Stdout = &mpv.output
	err := cmd.Start()
	if err != nil {
		slog.Error("mpv cmd start", "error", err)
		return err
	}
	mpv.cmd = cmd
	slog.Debug("mpv cmd started", "pid", mpv.cmd.Process.Pid)

	return nil
}

func (mpv *Mpv) Metadata() (*Metadata, error) {
	if mpv.cmd == nil {
		return nil, nil
	}

	output := mpv.output.String()
	slog.Debug("mpv", "output", output)
	errIx := strings.Index(output, errOut)
	if errIx >= 0 {
		errMsg := output[errIx:]
		nlIx := strings.Index(errMsg, "\n")
		if nlIx >= 0 {
			errMsg = errMsg[:nlIx]
		}
		errMsg = strings.TrimSpace(errMsg)
		return nil, errors.New(errMsg)
	}
	title := ""
	titleIx := strings.LastIndex(output, titleMsg)
	if titleIx >= 0 {
		title = output[titleIx+10:]
		nlIx := strings.Index(title, "\n")
		if nlIx >= 0 {
			title = title[:nlIx]
		}
	}
	title = strings.TrimSpace(title)

	res := Metadata{Title: title}
	return &res, nil
}

func (mpv *Mpv) Stop() error {
	mpv.output.Reset()

	if mpv.cmd == nil {
		slog.Debug("no current station playing")
		return nil
	}
	// err := m.cmd.Wait()
	// if err != nil {
	// 	return err
	// }

	if mpv.cmd.Process != nil {
		slog.Debug("killing process", "pid", mpv.cmd.Process.Pid)

		// err := m.cmd.Process.Kill()
		// if err != nil {
		// 	return err
		// }

		// pid := m.cmd.Process.Pid
		pid, err := syscall.Getpgid(mpv.cmd.Process.Pid)
		if err != nil {
			slog.Error("error getting process group", "pid", mpv.cmd.Process.Pid)
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
		mpv.cmd = nil
	}

	return nil
}
