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
	"sync/atomic"
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
	password
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
	password:    "password %s",
}

type Mpd struct {
	password   *string
	conn       net.Conn
	nowPlaying atomic.Bool
}

func New(ctx context.Context, host string, port int, password *string) (*Mpd, error) {
	p := &Mpd{password: password}
	conn, err := getConn(ctx, host, port)
	if err != nil {
		return nil, err
	}
	p.conn = conn

	_ = p.setPassword()

	return p, nil
}

const incorrectPass = "incorrect password"

var errIncorectPass = errors.New("incorrect MPD password")

func (m *Mpd) setPassword() error {
	if m.password == nil {
		return nil
	}
	cmd := fmt.Sprintf(cmds[password], *m.password)
	out, err := m.doCmd(cmd)
	if err != nil {
		return err
	} else if strings.Contains(out, incorrectPass) {
		return errIncorectPass
	}
	return nil
}

func getConn(ctx context.Context, host string, port int) (net.Conn, error) {
	var d net.Dialer
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := d.DialContext(ctx, "tcp", addr)
	slog.Info("mpd tcp", "address", addr, "err", err)
	if err != nil {
		addr = fmt.Sprintf("%s:%d", config.DefMpdHost, config.DefMpdPort)
		conn, err = d.DialContext(ctx, "tcp", addr)
		slog.Info("mpd tcp", "default address", addr, "err", err)
	}
	return conn, err
}

func (m *Mpd) Play(streamURL string) error {
	_, err := m.doCmd(cmds[clear])
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf(cmds[add], streamURL)
	_, err = m.doCmd(cmd)
	if err != nil {
		return err
	}

	_, err = m.doCmd(cmds[play])
	if err != nil {
		return err
	}

	m.nowPlaying.Store(true)
	return nil
}

func (m *Mpd) Pause(value bool) error {
	cmd := cmds[pause]
	_, err := m.doCmd(cmd)
	if err != nil {
		return err
	}

	m.nowPlaying.Store(!value)
	return nil
}

const (
	mpdRespOk    = "MPD Response: OK"
	setvolErrMsg = "Failed to set volume: ACK"
	volumeMsg    = "volume:"
)

var errNotPlaying = errors.New("a sound must be playing for MPD’s volume to be adjusted")

// SetVolume: a sound must be playing for MPD’s volume to be adjusted.
func (m *Mpd) SetVolume(value int) (int, error) {
	if !m.nowPlaying.Load() {
		return 0, errNotPlaying
	}
	if err := m.doSetvol(value); err != nil {
		return 0, err
	}
	return m.doGetvol()
}

func (m *Mpd) doSetvol(value int) error {
	cmd := fmt.Sprintf(cmds[setvol], value)
	out, err := m.doCmd(cmd)
	if err != nil {
		return fmt.Errorf("setvol cmd err: %w", err)
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
			return fmt.Errorf("setvol response err: %v", out)
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("setvol scanner err: %v", out)
	}
	return nil
}

func (m *Mpd) doGetvol() (int, error) {
	out, err := m.doCmd(cmds[getvol])
	if err != nil {
		return 0, fmt.Errorf("getvol cmd err: %w", err)
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		l := sc.Text()
		idx := strings.Index(l, volumeMsg)
		if idx == -1 {
			continue
		}
		volStr := strings.TrimSpace(l[len(volumeMsg):])
		parsedVolume, err := strconv.Atoi(volStr)
		if err != nil {
			return 0, fmt.Errorf("parse volume %s err: %v", out, err)
		}
		return parsedVolume, nil
	}
	if err := sc.Err(); err != nil {
		return 0, fmt.Errorf("getvol scanner err: %v", out)
	}
	return 0, fmt.Errorf("failed to parse volume: %s", out)
}

const titleMsg = "Title:"

func (m *Mpd) Metadata() *model.Metadata {
	out, err := m.doCmd(cmds[currentSong])
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

const elapsedMsg = "elapsed:"

func (m *Mpd) getElapsedSeconds() (int64, error) {
	out, err := m.doCmd(cmds[status])
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

const notSeekableMsg = "not seekable"

func (m *Mpd) Seek(amtSec int) *model.Metadata {
	elapsedSecs, err := m.getElapsedSeconds()
	if err != nil {
		return &model.Metadata{Err: err}
	}

	pos := elapsedSecs + int64(amtSec)
	if pos < 0 {
		pos = 0
	}
	cmd := fmt.Sprintf(cmds[seekcurr], pos)
	out, err := m.doCmd(cmd)
	if err != nil {
		return &model.Metadata{Err: err}
	} else if strings.Contains(strings.ToLower(out), notSeekableMsg) {
		return nil
	}

	return m.Metadata()
}

func (m *Mpd) Stop() error {
	_, err := m.doCmd(cmds[stop])
	if err != nil {
		return err
	}

	m.nowPlaying.Store(false)
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

	_, err = m.doCmd(cmds[clear])
	if err != nil {
		return err
	}

	return nil
}

func (m *Mpd) GetType() config.PlayerType {
	return config.MPD
}

const wrongPermission = "you don't have permission for"

var errWrongPermission = errors.New("MPD permission error")

func (m *Mpd) doCmd(cmd string) (string, error) {
	cmd += "\n"
	log := slog.With("method", "Mpd.doMpcCmd")
	doLog := true
	if strings.Contains(cmd, "password") {
		doLog = false
	}
	if doLog {
		log = log.With("cmd", cmd)
		log.Info("start")
		defer func() {
			log.Info("stop")
		}()
	}

	m.conn.SetDeadline(time.Now().Add(config.MpdConnTimeout))
	_, err := m.conn.Write([]byte(cmd))
	if err != nil {
		return "", fmt.Errorf("MPD write err: %w", err)
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
		return "", fmt.Errorf("MPD read error: %w", err)
	}
	resS := res.String()
	if doLog {
		log.Info(fmt.Sprintf("<<<\n%s\n", resS))
	}
	if strings.Contains(resS, wrongPermission) {
		return "", errWrongPermission
	}
	return resS, nil
}
