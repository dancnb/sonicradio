package player

import (
	"context"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/ffplay"
	"github.com/dancnb/sonicradio/player/model"
	"github.com/dancnb/sonicradio/player/mpv"
)

type Player struct {
	delegate   backendPlayer
	playerType config.PlayerType
}

type backendPlayer interface {
	GetType() config.PlayerType
	Play(url string) error
	Pause(value bool) error
	Stop() error
	SetVolume(value int) (int, error)
	Metadata() *model.Metadata
	Seek(amtSec int) *model.Metadata
	Close() error
}

func NewPlayer(ctx context.Context, cfg *config.Value) (*Player, error) {
	var delegate backendPlayer
	switch cfg.Player {
	case config.Mpv:
		mpvPlayer, err := mpv.NewMPVSocket(ctx)
		if err != nil {
			return nil, err
		}
		delegate = mpvPlayer
	case config.FFPlay:
		ffplayPlayer, err := ffplay.NewFFPlay(ctx)
		if err != nil {
			return nil, err
		}
		delegate = ffplayPlayer
	}

	vol := cfg.GetVolume()
	_, err := delegate.SetVolume(clampVolume(vol))
	if err != nil {
		return nil, err
	}

	return &Player{delegate: delegate, playerType: config.Mpv}, nil
}

func (p *Player) Play(url string) error {
	return p.delegate.Play(url)
}

func (p *Player) Pause(value bool) error {
	return p.delegate.Pause(value)
}

func (p *Player) Stop() error {
	return p.delegate.Stop()
}

func clampVolume(value int) int {
	if value < 0 {
		value = 0
	} else if value > 100 {
		value = 100
	}
	return value
}

func (p *Player) SetVolume(value int) (int, error) {
	return p.delegate.SetVolume(clampVolume(value))
}

func (p *Player) Metadata() *model.Metadata {
	return p.delegate.Metadata()
}

func (p *Player) Seek(amtSec int) *model.Metadata {
	return p.delegate.Seek(amtSec)
}

func (p *Player) Close() error {
	return p.delegate.Close()
}
