package mpd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
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
	currentSong
	status
	seekcurr
	setvol
	getvol
)

var cmds = map[command]string{
	play:        "play",
	pause:       "pause",
	clear:       "clear",
	add:         "add %s",
	stop:        "stop",
	currentSong: "currentsong",
	status:      "status",
	seekcurr:    "seekcur %d",
	setvol:      "setvol %d",
	getvol:      "getvol",
}

type Mpd struct {
	conn   net.Conn
	volume int
}

func New(ctx context.Context) (*Mpd, error) {
	p := &Mpd{}
	conn, err := getConn(ctx)
	if err != nil {
		return nil, err
	}
	p.conn = conn
	return p, nil
}

func getConn(ctx context.Context) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", "localhost:6600")
	return conn, err
}

func (m *Mpd) doCmd(cmd string) (string, error) {
	cmd += "\n"
	log := slog.With("method", "Mpd.doMpcCmd")
	log.Info("start", "args", cmd)
	defer func() {
		log.Info("stop", "args", cmd)
	}()

	m.conn.SetDeadline(time.Now().Add(config.MpdConnTimeout))
	_, err := m.conn.Write([]byte(cmd))
	if err != nil {
		return "", fmt.Errorf("write err: %w", err)
	}

	scanner := bufio.NewScanner(m.conn)
	m.conn.SetDeadline(time.Now().Add(config.VlcConnTimeout))
	var res strings.Builder
	for scanner.Scan() {
		l := scanner.Text()
		res.WriteString(l)
		res.WriteString("\n")
		m.conn.SetDeadline(time.Now().Add(config.VlcConnTimeout))
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
		return "", fmt.Errorf("scanner error: %w", err)
	}
	resS := res.String()
	return resS, nil
}

func (m *Mpd) GetType() config.PlayerType {
	return config.MPD
}

func (m *Mpd) Pause(value bool) error {
	log := slog.With("method", "Mpd.Pause")

	cmd := cmds[pause]
	b, err := m.doCmd(cmd)
	logCmdOutput(log, b)
	if err != nil {
		return err
	}

	return nil
}

func logCmdOutput(log *slog.Logger, output string) {
	log.Info(fmt.Sprintf("<<<\n%s\n", output))
}

const (
	mpdRespOk    = "MPD Response: OK"
	setvolErrMsg = "Failed to set volume: ACK"
	volumeMsg    = "volume:"
)

func (m *Mpd) SetVolume(value int) (int, error) {
	log := slog.With("method", "Mpd.SetVolume")

	cmd := fmt.Sprintf(cmds[setvol], value)
	out, err := m.doCmd(cmd)
	logCmdOutput(log, out)
	if err != nil {
		return m.volume, fmt.Errorf("setvol cmd err: %w", err)
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		l := sc.Text()
		idx := strings.Index(l, mpdRespOk)
		if idx != -1 {
			break
		}
		idx = strings.Index(l, setvolErrMsg)
		if idx != -1 {
			return m.volume, fmt.Errorf("setvol response err: %v", out)
		}
	}
	if err := sc.Err(); err != nil {
		return m.volume, fmt.Errorf("setvol scanner err: %v", out)
	}

	out, err = m.doCmd(cmds[getvol])
	logCmdOutput(log, out)
	if err != nil {
		return m.volume, fmt.Errorf("getvol cmd err: %w", err)
	}
	sc = bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		l := sc.Text()
		idx := strings.Index(l, volumeMsg)
		if idx == -1 {
			continue
		}
		volStr := strings.TrimSpace(l[len(volumeMsg):])
		volInt, err := strconv.Atoi(volStr)
		if err != nil {
			return m.volume, fmt.Errorf("parse volume %s err: %v", out, err)
		}
		m.volume = volInt
		return m.volume, nil
	}
	if err := sc.Err(); err != nil {
		return m.volume, fmt.Errorf("getvol scanner err: %v", out)
	}
	return m.volume, nil
}

const (
	titleMsg   = "Title:"
	elapsedMsg = "elapsed:"
)

func (m *Mpd) Metadata() *model.Metadata {
	log := slog.With("method", "Mpd.Metadata")

	out, err := m.doCmd(cmds[currentSong])
	logCmdOutput(log, out)
	if err != nil {
		return &model.Metadata{Err: fmt.Errorf("currentsong cmd err: %w", err)}
	}
	meta := &model.Metadata{}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		l := sc.Text()
		idx := strings.Index(l, titleMsg)
		if idx == -1 {
			continue
		}
		meta.Title = strings.TrimSpace(l[len(titleMsg):])
		break
	}
	if err := sc.Err(); err != nil {
		return &model.Metadata{Err: fmt.Errorf("currentsong scanner err: %w", err)}
	}

	intSecs, err := m.getElapsedSeconds()
	if err != nil {
		return &model.Metadata{Err: err}
	}
	meta.PlaybackTimeSec = &intSecs
	return meta
}

func (m *Mpd) getElapsedSeconds() (int64, error) {
	log := slog.With("method", "Mpd.getElapsedSeconds")
	out, err := m.doCmd(cmds[status])
	logCmdOutput(log, out)
	if err != nil {
		return -1, fmt.Errorf("status cmd err: %w", err)
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		l := sc.Text()
		idx := strings.Index(l, elapsedMsg)
		if idx == -1 {
			continue
		}
		elapsed := strings.TrimSpace(l[len(elapsedMsg):])
		f, err := strconv.ParseFloat(elapsed, 64)
		if err != nil {
			return -1, fmt.Errorf("parsed elapsed(%s) time err: %w", out, err)
		}
		intEl := int64(f)
		return intEl, nil
	}
	if err := sc.Err(); err != nil {
		return -1, fmt.Errorf("status scanner(%s) err: %w", out, err)
	}
	return -1, fmt.Errorf("could not parse elapsed time from: %s", out)
}

func (m *Mpd) Seek(amtSec int) *model.Metadata {
	log := slog.With("method", "Mpd.Seek")

	intSecs, err := m.getElapsedSeconds()
	if err != nil {
		return &model.Metadata{Err: err}
	}

	cmd := fmt.Sprintf(cmds[seekcurr], intSecs+int64(amtSec))
	b, err := m.doCmd(cmd)
	logCmdOutput(log, b)
	if err != nil {
		return &model.Metadata{Err: err}
	}

	return m.Metadata()
}

func (m *Mpd) Play(url string) error {
	log := slog.With("method", "Mpd.Play")

	b, err := m.doCmd(cmds[clear])
	logCmdOutput(log, b)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf(cmds[add], url)
	b, err = m.doCmd(cmd)
	logCmdOutput(log, b)
	if err != nil {
		return err
	}

	b, err = m.doCmd(cmds[play])
	logCmdOutput(log, b)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mpd) Stop() error {
	log := slog.With("method", "Mpd.Stop")

	b, err := m.doCmd(cmds[stop])
	logCmdOutput(log, b)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mpd) Close() (err error) {
	log := slog.With("method", "Mpd.Close")
	log.Info("stopping")
	defer func() {
		log.Info("stopped")
	}()

	defer func() {
		if m.conn != nil {
			if closeErr := m.conn.Close(); closeErr != nil {
				log.Error("mpd tcp connection close", "err", closeErr)
				if err == nil {
					err = closeErr
				}
			}
		}
	}()

	b, err := m.doCmd(cmds[clear])
	logCmdOutput(log, b)
	if err != nil {
		return err
	}

	return nil
}
