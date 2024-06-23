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
	url string
	cmd *exec.Cmd
}

func (mpv *Mpv) Play(url string) error {
	log := slog.With("method", "Mpv.Play")
	log.Info("playing url=" + url)
	if err := mpv.Stop(); err != nil {
		return err
	}

	args := slices.Clone(baseArgs)
	args = append(args, url)
	cmd := exec.Command(baseCmd, args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		log.Error("mpv cmd error", "error", cmd.Err.Error())
		return cmd.Err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = &bytes.Buffer{}
	err := cmd.Start()
	if err != nil {
		log.Error("mpv cmd start", "error", err)
		return err
	}
	mpv.cmd = cmd
	mpv.url = url
	log.Debug("mpv cmd started", "pid", mpv.cmd.Process.Pid)

	return nil
}

func (mpv *Mpv) Metadata() *Metadata {
	if mpv.cmd == nil || mpv.cmd.Stdout == nil {
		return nil
	}
	log := slog.With("method", "Mpv.Metadata")

	output := mpv.cmd.Stdout.(*bytes.Buffer).String()

	log.Debug("mpv", "output", output)
	errIx := strings.Index(output, errOut)
	if errIx >= 0 {
		errMsg := output[errIx:]
		nlIx := strings.Index(errMsg, "\n")
		if nlIx >= 0 {
			errMsg = errMsg[:nlIx]
		}
		errMsg = strings.TrimSpace(errMsg)
		return &Metadata{URL: mpv.url, Err: errors.New(errMsg)}
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

	return &Metadata{URL: mpv.url, Title: title}
}

func (mpv *Mpv) Stop() error {
	log := slog.With("method", "Mpv.Stop")

	mpv.url = ""
	if mpv.cmd == nil {
		log.Debug("no current station playing")
		return nil
	}
	cmd := *mpv.cmd
	mpv.cmd = nil
	cmd.Stdout = nil

	// err := cmd.Wait()
	// if err != nil {
	// 	return err
	// }

	if cmd.Process != nil {
		log.Debug("killing process", "pid", cmd.Process.Pid)

		// err := cmd.Process.Kill()
		// if err != nil {
		// 	return err
		// }

		// pid := cmd.Process.Pid
		pid, err := syscall.Getpgid(cmd.Process.Pid)
		if err != nil {
			log.Error("error getting process group", "pid", cmd.Process.Pid)
			return err
		}
		// err = syscall.Kill(-cmd.Process.Pid, syscall.SIGCHLD)
		// if err != nil {
		// 	log.Error("error killing process group", "pgid", pid)
		// 	return err
		// }
		err = syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			log.Error("error killing process children", "pgid", pid)
			return err
		}

		log.Debug("killed process group", "pgid", pid)
	}

	return nil
}
