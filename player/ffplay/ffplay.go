package ffplay

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
)

const (
	errOut = "Failed to"

	// titleMsg = "icy-title:"
	titleMsg = "StreamTitle"
)

var (
	baseArgs = []string{"-hide_banner", "-nodisp", "-loglevel", "verbose", "-autoexit", "-volume"}
	volArg   = "%d"
)

type FFPlay struct {
	url     string
	playing *exec.Cmd

	volume int
}

func NewFFPlay(ctx context.Context) (*FFPlay, error) {
	return &FFPlay{}, nil
}

func (f *FFPlay) GetType() config.PlayerType {
	return config.FFPlay
}

func (f *FFPlay) Play(url string) error {
	log := slog.With("method", "FFPlay.Play")
	log.Info("playing url=" + url)
	if err := f.Stop(); err != nil {
		return err
	}

	args := slices.Clone(baseArgs)
	args = append(args, fmt.Sprintf(volArg, f.volume))
	args = append(args, url)
	cmd := exec.Command(getBaseCmd(), args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		log.Error("ffplay cmd error", "error", cmd.Err.Error())
		return cmd.Err
	}
	cmd.Stderr = &bytes.Buffer{}
	err := cmd.Start()
	if err != nil {
		log.Error("ffplay cmd start", "error", err)
		return err
	}
	f.playing = cmd
	f.url = url
	log.Debug("ffplay cmd started", "pid", f.playing.Process.Pid)

	return nil
}

func (f *FFPlay) Pause(value bool) error {
	if value {
		return f.Stop()
	} else if f.url != "" {
		return f.Play(f.url)
	}
	return nil
}

func (f *FFPlay) Stop() error {
	log := slog.With("method", "FFPlay.Stop")
	if f.playing == nil {
		log.Debug("no current station playing")
		return nil
	}
	cmd := *f.playing
	f.playing = nil
	cmd.Stderr = nil
	return killProcess(cmd.Process, log)
}

func (f *FFPlay) SetVolume(value int) (int, error) {
	log := slog.With("method", "FFPlay.SetVolume")
	log.Info("volume", "value", value)
	if f.playing == nil {
		f.volume = value
	}
	return f.volume, nil
}

// TODO: error msg
// TODO: playback time
func (f *FFPlay) Metadata() *model.Metadata {
	if f.playing == nil || f.playing.Stderr == nil {
		return nil
	}
	log := slog.With("method", "FFPlay.Metadata")

	output := f.playing.Stderr.(*bytes.Buffer).String()
	log.Debug("FFPlay", "output", output)

	errIx := strings.Index(output, errOut)
	if errIx >= 0 {
		errMsg := output[errIx:]
		nlIx := strings.Index(errMsg, "\n")
		if nlIx >= 0 {
			errMsg = errMsg[:nlIx]
		}
		errMsg = strings.TrimSpace(errMsg)
		return &model.Metadata{Err: errors.New(errMsg)}
	}

	title := ""
	titleIx := strings.LastIndex(output, titleMsg)
	if titleIx >= 0 {
		title = output[titleIx+10:]
		nlIx := strings.Index(title, "\n")
		if nlIx >= 0 {
			title = title[:nlIx]
		}
		titleParts := strings.Split(title, ":")
		if len(titleParts) != 2 {
			return &model.Metadata{Err: fmt.Errorf("invalid title message %s", title)}
		}
		title = strings.TrimSpace(titleParts[1])
	}

	return &model.Metadata{Title: title}
}

func (f *FFPlay) Seek(amtSec int) *model.Metadata {
	return nil
}

func (f *FFPlay) Close() error {
	return nil
}
