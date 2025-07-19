package player

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/ffplay"
	"github.com/dancnb/sonicradio/player/model"
	"github.com/dancnb/sonicradio/player/mpd"
	"github.com/dancnb/sonicradio/player/mplayer"
	"github.com/dancnb/sonicradio/player/mpv"
	"github.com/dancnb/sonicradio/player/vlc"
)

type Player struct {
	delegate  backendPlayer
	available map[config.PlayerType]struct{}
}

type backendPlayer interface {
	GetType() config.PlayerType
	Play(url string) error
	Pause(value bool) error
	Stop() error

	// SetVolume:
	//
	//   - sets the volume to an absolute value in [0,100]
	//   - returns the set value and nil if succeeded, error if failed
	SetVolume(value int) (int, error)

	Metadata() *model.Metadata

	// Seek:
	//
	//   - seek by a +/- amount of seconds,
	//   - returns the metadata for the new playback position if succeeded, metadata with error if failed
	Seek(amtSec int) *model.Metadata

	Close() error
}

func NewPlayer(ctx context.Context, cfg *config.Value) (*Player, error) {
	p := new(Player)
	err := p.checkPlayerType(cfg)
	if err != nil {
		return nil, err
	}

	vol := cfg.GetVolume()
	switch cfg.Player {
	case config.Mpv:
		mpvPlayer, err := mpv.NewMPVSocket(ctx)
		if err != nil {
			return nil, err
		}
		p.delegate = mpvPlayer
	case config.FFPlay:
		ffplayPlayer, err := ffplay.NewFFPlay(ctx)
		if err != nil {
			return nil, err
		}
		p.delegate = ffplayPlayer
	case config.Vlc:
		vlcPlayer, err := vlc.NewVlc(ctx)
		if err != nil {
			return nil, err
		}
		p.delegate = vlcPlayer
	case config.MPlayer:
		mplayer, err := mplayer.New(ctx, vol)
		if err != nil {
			return nil, err
		}
		p.delegate = mplayer
	case config.MPD:
		mpdp, err := mpd.New(ctx, cfg.MpdHost, cfg.MpdPort, cfg.GetMpdPassword())
		if err != nil {
			return nil, err
		}
		p.delegate = mpdp
	}

	_, err = p.delegate.SetVolume(clampVolume(vol))
	if err != nil && cfg.Player != config.MPD {
		return nil, err
	}

	return p, nil
}

var errNoPlayerAvailable = errors.New("No available player found. Must have at least one of the following in PATH: mpv, ffplay, vlc.")

func (p *Player) checkPlayerType(cfg *config.Value) error {
	p.available = make(map[config.PlayerType]struct{}, len(config.Players))
	var firstAvailable *config.PlayerType
	for _, v := range config.Players {
		if ok := checkAvailablePlayer(v); !ok {
			continue
		}
		if firstAvailable == nil {
			firstAvailable = &v
		}
		p.available[v] = struct{}{}
	}
	if len(p.available) == 0 {
		return errNoPlayerAvailable
	}
	if _, ok := p.available[cfg.Player]; !ok {
		cfg.Player = *firstAvailable
	}
	slog.Info("Player.checkPlayerType", "value", cfg.Player)
	return nil
}

var baseCmds = map[config.PlayerType]func() string{
	config.Mpv:     mpv.GetBaseCmd,
	config.FFPlay:  ffplay.GetBaseCmd,
	config.Vlc:     vlc.GetBaseCmd,
	config.MPlayer: mplayer.GetBaseCmd,
	config.MPD:     mpd.GetBaseCmd,
}

func checkAvailablePlayer(p config.PlayerType) bool {
	baseCmdFn, ok := baseCmds[p]
	if !ok {
		return false
	}
	baseCmd := baseCmdFn()
	path, err := exec.LookPath(baseCmd)
	slog.Info("checkAvailablePlayer", "cmd", baseCmd, "path", path, "err", err)
	if err != nil && !errors.Is(err, exec.ErrDot) {
		return false
	}
	return true
}

func (p *Player) PlayerTypes() []config.PlayerType {
	var res []config.PlayerType
	for _, v := range config.Players {
		if _, ok := p.available[v]; ok {
			res = append(res, v)
		}
	}
	return res
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

// SetVolume:
//
//   - sets the volume to an absolute value in [0,100]
//   - returns the set value and nil if succeeded, or an irrelevant value and error if failed
func (p *Player) SetVolume(value int) (int, error) {
	return p.delegate.SetVolume(clampVolume(value))
}

func (p *Player) Metadata() *model.Metadata {
	return p.delegate.Metadata()
}

// Seek:
//
//   - seek by a +/- amount of seconds,
//   - returns the metadata for the new playback position if succeeded, metadata with error if failed
func (p *Player) Seek(amtSec int) *model.Metadata {
	return p.delegate.Seek(amtSec)
}

func (p *Player) Close() error {
	return p.delegate.Close()
}
