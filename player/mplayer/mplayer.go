package mplayer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"slices"
	"strings"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
	playerutils "github.com/dancnb/sonicradio/player/utils"
)

var (
	volArg   = "-volume"
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
	// getTime
	// seek
	volume // set volume abs percentage
)

var cmds = map[command]string{
	play:  "loadfile %s",
	pause: "pause",
	stop:  "pausing_keep_force stop",
	quit:  "quit",
	// getTime: "pausing_keep_force get_time", //alt "get_time_pos"
	// seek:    "seek %d",
	volume: "pausing_keep_force volume %d 1",
}

type Mplayer struct {
	cmd *exec.Cmd
	wc  io.WriteCloser
	rc  io.ReadCloser

	title *string
	pt    *playerutils.PlaybackTime
}

func New(ctx context.Context, volume int) (*Mplayer, error) {
	p := &Mplayer{
		pt: &playerutils.PlaybackTime{},
	}
	err := p.getCmd(ctx, volume)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (m *Mplayer) getCmd(ctx context.Context, volume int) error {
	log := slog.With("method", "Mplayer.getCmd")
	args := []string{volArg, fmt.Sprintf("%d", volume)}
	args = append(args, slices.Clone(baseArgs)...)
	cmd := exec.CommandContext(ctx, GetBaseCmd(), args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		log.Error("mplayer cmd error", "error", cmd.Err.Error())
		return cmd.Err
	}

	wc, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	rc, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		log.Error("mplayer cmd start", "error", err)
		return err
	}
	m.cmd = cmd
	m.wc = wc
	m.rc = rc
	log.Info("mplayer cmd started", "pid", cmd.Process.Pid)

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

}

func (m *Mplayer) parseOutputLine(logger *slog.Logger, output string) {
	logger.Info("<<<< " + output)

	startIdx := strings.Index(output, titleMsg)
	if startIdx != -1 {
		titleS := output[startIdx+len(titleMsg):]
		endIdx := strings.Index(titleS, "'")
		if endIdx != -1 {
			titleS = titleS[:endIdx]
			m.title = &titleS
		}
	}
}

func (m *Mplayer) GetType() config.PlayerType {
	return config.MPlayer
}

func (m *Mplayer) Pause(value bool) error {
	_, err := m.doCommand(cmds[pause])
	if err == nil {
		if value {
			m.pt.PausePlayTime()
		} else {
			m.pt.ResumePlayTime()
		}
	}
	return err
}

func (m *Mplayer) SetVolume(value int) (int, error) {
	cmd := fmt.Sprintf(cmds[volume], value)
	_, err := m.doCommand(cmd)
	return value, err
}

func (m *Mplayer) Metadata() *model.Metadata {
	metadata := &model.Metadata{PlaybackTimeSec: m.pt.GetPlayTime()}
	if m.title != nil {
		metadata.Title = *m.title
	}
	return metadata
}

func (m *Mplayer) Seek(amtSec int) *model.Metadata {
	return nil
}

func (m *Mplayer) Play(url string) error {
	m.title = nil

	if err := m.Stop(); err != nil {
		return err
	}

	cmd := fmt.Sprintf(cmds[play], url)
	_, err := m.doCommand(cmd)
	if err == nil {
		m.pt.ResetPlayTime()
	}

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
	logger.Info(">>>> " + cmd)
	cmd = cmd + "\n"
	_, err := io.WriteString(m.wc, cmd)
	if err != nil {
		logger.Error("write", "err", err)
		return "", err
	}
	// if err := m.wc.(*os.File).Sync(); err != nil {
	// 	logger.Error("sync", "err", err)
	// }

	return "", nil
}
