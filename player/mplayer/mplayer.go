package mplayer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
	playerutils "github.com/dancnb/sonicradio/player/utils"
)

var (
	baseArgs = []string{
		// "-cache", "20480",
		// "-cache-min", "50",
		// "-cache-seek-min", "20",
		"-idle", "-slave", "-quiet",
	}
)

type command uint8

const (
	play command = iota
	pause
	stop
	quit
	getTime
	seek
	volume // set volume abs percentage
)

var cmds = map[command]string{
	play:    "loadfile %s",
	pause:   "pause",
	stop:    "stop",
	quit:    "quit",
	getTime: "get_time", //alt "get_time_pos"
	seek:    "seek %d",
	volume:  "volume %d 1",
}

type Mplayer struct {
	cmd  *exec.Cmd
	done chan struct{}
	wc   io.WriteCloser
	rc   io.ReadCloser

	title *string
	time  *int64
}

func New(ctx context.Context) (*Mplayer, error) {
	p := &Mplayer{
		done: make(chan struct{}),
	}
	err := p.getCmd(ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (m *Mplayer) getCmd(ctx context.Context) error {
	log := slog.With("method", "Mplayer.getCmd")
	args := slices.Clone(baseArgs)
	cmd := exec.CommandContext(ctx, GetBaseCmd(), args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		log.Error("mpv cmd error", "error", cmd.Err.Error())
		return cmd.Err
	}

	wc, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	// rc, err := cmd.StderrPipe()
	rc, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		log.Error("mpv cmd start", "error", err)
		return err
	}
	m.cmd = cmd
	m.wc = wc
	m.rc = rc
	log.Info("mpv cmd started", "pid", cmd.Process.Pid)

	go m.readOutput(ctx)

	return nil
}

const (
	titleMsg = "StreamTitle='"
	timeMsg  = "ANS_TIME_POSITION="
)

func (m *Mplayer) readOutput(ctx context.Context) {
	logger := slog.With("method", "Mplayer.readOutput")
	sc := bufio.NewScanner(m.rc)

	for sc.Scan() {
		l := sc.Text()
		m.parseOutputLine(logger, l)
	}
	logger.Info("scanner finished")
	if err := sc.Err(); err != nil {
		logger.Info("scanner error", "", err)
	}

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		//drain output
	// 		for sc.Scan() {
	// 			l := sc.Text()
	// 			m.parseOutputLine(logger, l)
	// 		}
	// 		logger.Info("after drain scanner finished")
	// 		if err := sc.Err(); err != nil {
	// 			logger.Info("after drain scanner error", "", err)
	// 		}

	// 		return
	// 		// m.done <- struct{}{}
	// 	default:
	// 		if sc.Scan() {
	// 			l := sc.Text()
	// 			m.parseOutputLine(logger, l)
	// 		} else {
	// 			logger.Info("loop scanner finished")
	// 			if err := sc.Err(); err != nil {
	// 				logger.Info("loop scanner error", "", err)
	// 			}
	// 			return
	// 		}
	// 	}
	// }
}

func (m *Mplayer) parseOutputLine(logger *slog.Logger, output string) {
	logger.Info(output)
	if strings.TrimSpace(output) == "" {
		logger.Info(">>>>>><<<<<<<<")
	}

	startIdx := strings.Index(output, titleMsg)
	if startIdx != -1 {
		titleS := output[startIdx+len(titleMsg):]
		endIdx := strings.Index(titleS, "'")
		if endIdx != -1 {
			titleS = titleS[:endIdx]
			m.title = &titleS
		}
	}

	startIdx = strings.Index(output, timeMsg)
	if startIdx != -1 {
		timeS := output[startIdx+len(timeMsg):]
		t, err := strconv.ParseFloat(timeS, 64)
		if err != nil {
			logger.Error("parse time value", "err", err)
		} else {
			intT := int64(t)
			m.time = &intT
		}
	}
}

func (m *Mplayer) GetType() config.PlayerType {
	return config.MPlayer
}

func (m *Mplayer) Pause(value bool) error {
	_, err := m.doCommand(cmds[pause])
	return err
}

func (m *Mplayer) SetVolume(value int) (int, error) {
	cmd := fmt.Sprintf(cmds[volume], value)
	_, err := m.doCommand(cmd)
	return value, err
}

func (m *Mplayer) Metadata() *model.Metadata {
	if m.title == nil {
		return nil
	}
	cmd := cmds[getTime]
	_, err := m.doCommand(cmd)
	if err != nil {
		return nil
	}
	return &model.Metadata{Title: *m.title, PlaybackTimeSec: m.time}
}

func (m *Mplayer) Seek(amtSec int) *model.Metadata {
	return nil
}

func (m *Mplayer) Play(url string) error {
	m.title = nil
	m.time = nil

	if err := m.Stop(); err != nil {
		return err
	}

	cmd := fmt.Sprintf(cmds[play], url)
	_, err := m.doCommand(cmd)
	return err
}

func (m *Mplayer) Stop() error {
	_, err := m.doCommand(cmds[stop])
	return err
}

func (m *Mplayer) Close() (err error) {
	log := slog.With("method", "Mplayer.Close")

	defer func() {
		if m.wc != nil {
			if closeErr := m.wc.Close(); closeErr != nil {
				log.Error("Mplayer WriteCloser close", "err", closeErr)
				if err == nil {
					err = closeErr
				}
			}
		}
		if m.rc != nil {
			if closeErr := m.rc.Close(); closeErr != nil {
				log.Error("Mplayer ReadCloser close", "err", closeErr)
				if err == nil {
					err = closeErr
				}
			}
		}
	}()

	_, err = m.doCommand(cmds[quit])
	if err != nil {
		return err
	}
	// <-m.done

	if m.cmd != nil {
		if waitErr := m.cmd.Wait(); waitErr != nil {
			log.Error("Mplayer cmd wait", "err", waitErr)
			err = waitErr
		}
		if killErr := playerutils.KillProcess(m.cmd.Process, log); killErr != nil {
			log.Error("Mplayer cmd kill", "err", killErr)
		}
	}
	return err
}

func (m *Mplayer) doCommand(cmd string) (string, error) {
	logger := slog.With("method", "Mplayer.doCommand")
	logger.Info(cmd)
	cmd = cmd + "\n"
	_, err := io.WriteString(m.wc, cmd)
	if err != nil {
		logger.Error("write", "err", err)
		return "", err
	}
	if err := m.wc.(*os.File).Sync(); err != nil {
		logger.Error("sync", "err", err)
	}

	return "", nil
}
