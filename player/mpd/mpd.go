package mpd

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
)

var ()

type command uint8

const (
	play command = iota
	pause
	clear
	add
	stop
	status
	seek
	volume
	// quit
)

var cmds = map[command][]string{
	play:   {"play"},
	pause:  {"pause"},
	clear:  {"clear"},
	add:    {"add"},
	stop:   {"stop"},
	status: {"-f", "%artist% %title%", "status"},
	seek:   {"seek"},
	volume: {"volume"},
	// quit:   "",
}

type Mpd struct {
	volume int
}

func New(ctx context.Context) (*Mpd, error) {
	p := &Mpd{}
	return p, nil
}

var mpcTimeout = 2 * time.Second

// var mpcTimeout = 20 * time.Minute

func (m *Mpd) doMpcCmd(args []string) ([]byte, error) {
	log := slog.With("method", "Mpd.doMpcCmd")
	log.Info("start", "args", args)
	defer func() {
		log.Info("stop", "args", args)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), mpcTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, GetBaseCmd(), args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		log.Error("cmd error", "args", args, "err", cmd.Err)
		return nil, cmd.Err
	}

	cmd.Stdout = &bytes.Buffer{}

	err := cmd.Run()
	if err != nil {
		log.Error("cmd Run", "args", args, "err", err)
		return nil, err
	}

	b := cmd.Stdout.(*bytes.Buffer).Bytes()

	return b, err
}

func (m *Mpd) GetType() config.PlayerType {
	return config.MPD
}

func (m *Mpd) Pause(value bool) error {
	log := slog.With("method", "Mpd.Pause")

	args := cmds[pause]
	if !value {
		args = cmds[play]
	}
	b, err := m.doMpcCmd(args)
	if err != nil {
		return err
	}
	logCmdOutput(log, b)

	return nil
}

func logCmdOutput(log *slog.Logger, b []byte) {
	log.Info(fmt.Sprintf("<<<\n%s\n", b))
}

func (m *Mpd) SetVolume(value int) (int, error) {
	log := slog.With("method", "Mpd.SetVolume")

	args := cmds[volume]
	args = append(args, fmt.Sprintf("%d", value))
	b, err := m.doMpcCmd(args)
	if err != nil {
		return m.volume, err
	}
	logCmdOutput(log, b)

	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		l := sc.Text()
		slog.Debug(l)
		if !strings.Contains(l, "volume:") {
			continue
		} else if strings.Contains(l, "volume: n/a") {
			break
		}
		parts := strings.Fields(l)
		v, err := parseVolume(parts[0])
		if err != nil {
			return m.volume, err
		}
		m.volume = v
	}

	return m.volume, nil
}

func parseVolume(input string) (int, error) {
	trimmed := strings.TrimPrefix(input, "volume:")
	trimmed = strings.TrimSuffix(trimmed, "%")
	volume, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("invalid volume value: %s", input)
	}
	return volume, nil
}

func (m *Mpd) Metadata() *model.Metadata {
	log := slog.With("method", "Mpd.Metadata")

	b, err := m.doMpcCmd(cmds[status])
	if err != nil {
		return &model.Metadata{Err: err}
	}
	logCmdOutput(log, b)

	metadata := &model.Metadata{}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		if !sc.Scan() {
			break
		}
		l := sc.Text()
		if i == 0 {
			metadata.Title = strings.TrimSpace(l)
		} else if i == 1 {
			parts := strings.Fields(l)
			if len(parts) != 4 {
				break
			}
			parts = strings.Split(parts[2], "/")
			if len(parts) != 2 {
				break
			}
			d, err := parseDuration(parts[0])
			if err != nil {
				break
			}
			intSec := int64(d.Seconds())
			metadata.PlaybackTimeSec = &intSec
		}
	}

	return metadata
}

func parseDuration(input string) (time.Duration, error) {
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid duration format: %s", input)
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %s", parts[0])
	}

	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %s", parts[1])
	}

	return time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

func (m *Mpd) Seek(amtSec int) *model.Metadata {
	log := slog.With("method", "Mpd.Seek")

	args := cmds[seek]
	args = append(args, fmt.Sprintf("%d", amtSec))
	b, err := m.doMpcCmd(args)
	if err != nil {
		return &model.Metadata{Err: err}
	}
	logCmdOutput(log, b)

	return m.Metadata()
}

func (m *Mpd) Play(url string) error {
	log := slog.With("method", "Mpd.Play")

	b, err := m.doMpcCmd(cmds[clear])
	if err != nil {
		return err
	}
	logCmdOutput(log, b)

	args := cmds[add]
	args = append(args, url)
	b, err = m.doMpcCmd(args)
	if err != nil {
		return err
	}
	logCmdOutput(log, b)

	b, err = m.doMpcCmd(cmds[play])
	if err != nil {
		return err
	}
	logCmdOutput(log, b)

	return nil
}

func (m *Mpd) Stop() error {
	log := slog.With("method", "Mpd.Stop")

	b, err := m.doMpcCmd(cmds[stop])
	if err != nil {
		return err
	}
	logCmdOutput(log, b)

	return nil
}

func (m *Mpd) Close() (err error) {
	log := slog.With("method", "Mpd.Close")

	b, err := m.doMpcCmd(cmds[clear])
	if err != nil {
		return err
	}
	logCmdOutput(log, b)

	return nil
}
