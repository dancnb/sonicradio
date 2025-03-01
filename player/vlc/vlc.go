package vlc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
)

// var baseArgs = []string{"-I rc", "--no-video", "--volume-step 12.8", "--gain 1.0"}

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
)

var cmds = map[vlcRcCmd]string{
	add:      `add %s`,
	play:     `play`,
	stop:     `stop`,
	pause:    `pause`,
	unpause:  `pause`,
	volume:   `volume %f`,
	metadata: `info`,
	// mediaTitle:   `["get_property", "media-title"]`,
	playbackTime: `get_time`,
	seek:         `seek %d`,
	quit:         `quit`,
}

func NewVlc(ctx context.Context) (*Vlc, error) {
	p := &Vlc{}

	conn, err := getConn(ctx)
	if err != nil {
		return nil, err
	}
	p.conn = conn
	return p, nil
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
	cmd = cmds[play]
	_, err = v.doRequest(cmd)
	if err != nil {
		return err
	}
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

	quitCmd := cmds[quit]
	_, err = v.doRequest(quitCmd)
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

	v.conn.SetDeadline(time.Now().Add(config.MpvIpcConnTimeout))
	buff := make([]byte, 4096)
	_, err = v.conn.Read(buff)
	if err != nil {
		return "", fmt.Errorf("vlc read err: %w", err)
	}
	res := string(buff)
	log.Debug("vlc", "response", res)

	return res, nil
}
