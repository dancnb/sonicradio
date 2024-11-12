package player

import (
	"context"

	"github.com/dancnb/sonicradio/config"
)

type Player struct {
	mpv *MpvSocket
}

func NewPlayer(ctx context.Context, cfg config.Value) (*Player, error) {
	mpv, err := NewMPVSocket(ctx)
	if err != nil {
		return nil, err
	}
	vol := cfg.GetVolume()
	_, err = mpv.SetVolume(vol)
	if err != nil {
		return nil, err
	}
	return &Player{mpv}, nil
}

type Metadata struct {
	Title           string
	PlaybackTimeSec *int64
	Err             error
}

func (p *Player) Play(url string) error {
	return p.mpv.Play(url)
}

func (p *Player) Pause(value bool) error {
	return p.mpv.Pause(value)
}

func (p *Player) Stop() error {
	return p.mpv.Stop()
}

func (p *Player) SetVolume(value int) (int, error) {
	return p.mpv.SetVolume(value)
}

func (p *Player) Metadata() *Metadata {
	return p.mpv.Metadata()
}

func (p *Player) Seek(amtSec int) *Metadata {
	return p.mpv.Seek(amtSec)
}

func (p *Player) Close() error {
	return p.mpv.Close()
}
