package vlc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
)

// var baseArgs = []string{"-I rc", "--no-video", "--volume-step 12.8", "--gain 1.0"}
// var baseArgs = []string{"-I", "rc", "--rc-fake-tty", "--volume-step", "12.8", "--gain", "1.0", "--no-video", "--rc-host"}
var baseArgs = []string{"-I", "rc", "--volume-step", "12.8", "--gain", "1.0", "--no-video", "--rc-host"}

type Vlc struct {
	conn net.Conn
}

type vlcRcCmd uint8

const (
	add vlcRcCmd = iota
	play
	stop
	pause
	unpause
	volume
	metadata
	mediaTitle
	playbackTime
	seek
	quit
	shutdown
)

var cmds = map[vlcRcCmd]string{
	add:          "add %s\n",
	play:         "play\n",
	stop:         "stop\n",
	pause:        "pause\n",
	unpause:      "pause\n",
	volume:       "volume %f\n",
	metadata:     "info\n",
	playbackTime: "get_time\n",
	seek:         "seek %d\n",
	// quit:         "quit\n", // not good
	shutdown: "shutdown\n",
}

func NewVlc(ctx context.Context) (*Vlc, error) {
	p := &Vlc{}

	port, err := p.getAvailablePort()
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("localhost:%s", port)
	_, err = p.vlcCmd(ctx, addr)
	if err != nil {
		return nil, err
	}
	// mpv.cmd = cmd

	start := time.Now()
loop:
	for {
		select {
		case <-ctx.Done():
			return nil, ErrCtxCancel
		case <-time.After(socketTimeout):
			return nil, ErrSocketFileTimeout
		default:
			conn, err := getConn(ctx, addr)
			if err != nil {
				time.Sleep(socketSleepRetry)
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
	// cmd = cmds[play]
	// _, err = v.doRequest(cmd)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (v *Vlc) Pause(value bool) error {
	return nil
}

func (v *Vlc) Stop() error {
	return nil
}

func (v *Vlc) SetVolume(value int) (int, error) {
	// v := float64(value) * 2.56
	// f.cmd.Stderr
	return value, nil
}

func (v *Vlc) Metadata() *model.Metadata {
	return nil
}

func (v *Vlc) Seek(amtSec int) *model.Metadata {
	return nil
}

// TODO: also kill command?
func (v *Vlc) Close() (err error) {
	log := slog.With("method", "Vlc.Close")
	log.Info("stopping")

	defer func() {
		if v.conn == nil {
			return
		}
		closeErr := v.conn.Close()
		if closeErr != nil && err == nil {
			log.Error("vlc connection close", "err", closeErr)
			err = closeErr
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

	// v.conn.SetDeadline(time.Now().Add(config.VlcConnTimeout))
	// buff := make([]byte, 4096)
	// _, err = v.conn.Read(buff)
	// if err != nil {
	// 	return "", fmt.Errorf("vlc read err: %w", err)
	// }
	// res := string(buff)
	// log.Debug("vlc", "response", res)
	// return res, nil

	// scanner
	//
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
	// if err := scanner.Err(); err != nil {
	// 	return "", fmt.Errorf("scanner error: %w", err)
	// }
	return res.String(), nil

}
