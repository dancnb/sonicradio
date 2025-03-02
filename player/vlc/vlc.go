package vlc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
	playerutils "github.com/dancnb/sonicradio/player/utils"
)

// var baseArgs = []string{"-I rc", "--no-video", "--volume-step 12.8", "--gain 1.0"}
// var baseArgs = []string{"-I", "rc", "--rc-fake-tty", "--volume-step", "12.8", "--gain", "1.0", "--no-video", "--rc-host"}
var (
	socketTimeout    = time.Second * 2
	socketSleepRetry = time.Millisecond * 10

	ErrCtxCancel         = errors.New("context canceled")
	ErrSocketFileTimeout = errors.New("vlc socket file timeout")
	ErrNoMetadata        = errors.New("no metadata")
)

type Vlc struct {
	conn net.Conn
	cmd  *exec.Cmd
}

type vlcRcCmd uint8

const (
	add vlcRcCmd = iota
	play
	stop
	pause
	volume
	info
	mediaTitle
	getTime
	seek
	quit
	shutdown
)

var cmds = map[vlcRcCmd]string{
	add:      "add %s\n",
	play:     "play\n",
	stop:     "stop\n",
	pause:    "pause\n",
	volume:   "volume %f\n",
	info:     "info\n",
	getTime:  "get_time\n",
	seek:     "seek %d\n",
	quit:     "quit\n", // not good
	shutdown: "shutdown\n",
}

func NewVlc(ctx context.Context) (*Vlc, error) {
	p := &Vlc{}

	port, err := p.getAvailablePort()
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("localhost:%s", port)
	cmd, err := p.vlcCmd(ctx, addr)
	if err != nil {
		return nil, err
	}
	p.cmd = cmd

	start := time.Now()
loop:
	for {
		select {
		case <-ctx.Done():
			return nil, ErrCtxCancel
		case <-time.After(socketTimeout):
			return nil, ErrSocketFileTimeout
		default:
			conn, connErr := getConn(ctx, addr)
			if connErr != nil || conn == nil {
				time.Sleep(socketSleepRetry)
				continue
			}
			p.conn = conn
			break loop
		}
	}
	slog.Info(fmt.Sprintf("vlc tcp connection created after %v", time.Since(start)))

	return p, nil
}

func (v *Vlc) getAvailablePort() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer func() { _ = l.Close() }()
	addr := l.Addr().String()
	_, p, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	return p, nil
}

func (v *Vlc) vlcCmd(ctx context.Context, addr string) (*exec.Cmd, error) {
	log := slog.With("method", "vlcCmd")
	args := slices.Clone(baseArgs)
	args = append(args, addr)
	cmd := exec.CommandContext(ctx, GetBaseCmd(), args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		log.Error("vlc cmd error", "error", cmd.Err.Error())
		return nil, cmd.Err
	}
	err := cmd.Start()
	if err != nil {
		log.Error("vlc cmd start", "error", err)
		return nil, err
	}
	log.Info("vlc cmd started", "pid", cmd.Process.Pid)
	return cmd, nil
}

func getConn(ctx context.Context, addr string) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	return conn, err
}

func (v *Vlc) GetType() config.PlayerType {
	return config.Vlc
}

func (v *Vlc) Play(url string) error {
	cmd := fmt.Sprintf(cmds[add], url)
	_, err := v.doRequest(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (v *Vlc) Pause(value bool) error {
	cmd := cmds[pause]
	_, err := v.doRequest(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (v *Vlc) Stop() error {
	cmd := cmds[stop]
	_, err := v.doRequest(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (v *Vlc) SetVolume(value int) (int, error) {
	fVal := float64(value) * 2.56
	cmd := fmt.Sprintf(cmds[volume], fVal)
	_, err := v.doRequest(cmd)
	return value, err
}

const nowPlayingText = "now_playing:"

func (v *Vlc) Metadata() *model.Metadata {
	cmd := cmds[info]
	res, err := v.doRequest(cmd)
	if err != nil {
		return &model.Metadata{Err: err}
	}
	title := ""
	sc := bufio.NewScanner(strings.NewReader(res))
	for sc.Scan() {
		l := sc.Text()
		idx := strings.Index(l, nowPlayingText)
		if idx == -1 {
			continue
		}
		title = l[idx+len(nowPlayingText):]
		title = strings.TrimSpace(title)
		break
	}
	m := &model.Metadata{Title: title}

	cmd = cmds[getTime]
	res, err = v.doRequest(cmd)
	if err != nil {
		return m
	}
	sc = bufio.NewScanner(strings.NewReader(res))
	for sc.Scan() {
		l := sc.Text()
		l = strings.TrimSpace(l)
		intV, err := strconv.Atoi(l)
		if err != nil {
			continue
		}
		int64V := int64(intV)
		m.PlaybackTimeSec = &int64V
		break
	}
	return m
}

func (v *Vlc) Seek(amtSec int) *model.Metadata {
	return nil
}

func (v *Vlc) Close() (err error) {
	log := slog.With("method", "Vlc.Close")
	log.Info("stopping")

	defer func() {
		if v.conn != nil {
			closeErr := v.conn.Close()
			if closeErr != nil && err == nil {
				log.Error("vlc connection close", "err", closeErr)
				err = closeErr
			}
		}
		if v.cmd != nil {
			if killErr := playerutils.KillProcess(v.cmd.Process, log); killErr != nil {
				log.Error("vlc cmd kill", "err", killErr)
				err = killErr
			}
		}
	}()

	cmd := cmds[shutdown]
	_, err = v.doRequest(cmd)
	return err
}

func (v *Vlc) doRequest(cmd string) (string, error) {
	log := slog.With("method", "Vlc.doRequest")
	log.Info("vlc", "cmd", cmd)

	v.conn.SetDeadline(time.Now().Add(config.VlcConnTimeout))
	_, err := v.conn.Write([]byte(cmd))
	if err != nil {
		return "", fmt.Errorf("vlc write err: %w", err)
	}

	scanner := bufio.NewScanner(v.conn)
	v.conn.SetDeadline(time.Now().Add(config.VlcConnTimeout))
	var res strings.Builder
	for scanner.Scan() {
		l := scanner.Text()
		res.WriteString(l)
		res.WriteString("\n")
		log.Info(fmt.Sprintf("resp=%s", l))
		if l == "> " {
			break
		}
		v.conn.SetDeadline(time.Now().Add(config.VlcConnTimeout))
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
		return "", fmt.Errorf("scanner error: %w", err)
	}
	return res.String(), nil

}
