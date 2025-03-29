package mpd

import (
	"context"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
)

var ()

type command uint8

const (
	play command = iota
	pause
	stop
	quit
	status
	seek
	volume
)

var cmds = map[command]string{}

type Mpd struct {
}

func New(ctx context.Context) (*Mpd, error) {
	p := &Mpd{}
	return p, nil
}

/* func (m *Mpd) getCmd(ctx context.Context, volume int) error {
	// log := slog.With("method", "Mpd.getCmd")

	return nil
} */

func (m *Mpd) GetType() config.PlayerType {
	return config.MPD
}

func (m *Mpd) Pause(value bool) error {
	return nil
}

func (m *Mpd) SetVolume(value int) (int, error) {
	return value, nil
}

func (m *Mpd) Metadata() *model.Metadata {
	return nil
}

func (m *Mpd) Seek(amtSec int) *model.Metadata {
	return nil
}

func (m *Mpd) Play(url string) error {
	return nil
}

func (m *Mpd) Stop() error {
	return nil
}

func (m *Mpd) Close() (err error) {
	// log := slog.With("method", "Mpd.Close")

	return nil
}
