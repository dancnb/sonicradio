package internal

import (
	"context"
	"log/slog"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
)

type Internal struct {
	volume int
	cfg    *config.InternalPlayer

	// streamer
	cancelFn     context.CancelFunc
	buffStreamer *bufferedStreamer
}

func New(ctx context.Context, volume int, cfg *config.InternalPlayer) *Internal {
	return &Internal{
		volume: volume,
		cfg:    cfg,
	}
}

func (i *Internal) Play(url string) error {
	log := slog.With("caller", "Internal.Play", "url", url)
	log.Info("start")
	defer func() { log.Info("end") }()

	i.Stop()

	var ctx context.Context
	ctx, cancelFn := context.WithCancel(context.Background())
	buffStreamer, err := newBufferedStreamer(ctx, url, i.volume, i.cfg.BufferSeconds)
	if err != nil {
		slog.Info("newBufferedStreamer", "err", err.Error())
		cancelFn()
		return err
	}
	i.buffStreamer = buffStreamer
	i.cancelFn = cancelFn
	return nil
}

func (i *Internal) Pause(value bool) error {
	i.buffStreamer.togglePause()
	return nil
}

func (i *Internal) Stop() error {
	log := slog.With("caller", "Internal.Stop")
	log.Info("start")
	defer func() { log.Info("end") }()

	if i.cancelFn != nil {
		i.cancelFn()
		i.buffStreamer.wg.Wait()
	}
	return nil
}

func (i *Internal) SetVolume(value int) (int, error) {
	i.volume = value
	i.buffStreamer.setVolumeFromPercentage(value)
	return value, nil
}

func (i *Internal) Metadata() *model.Metadata {
	if i.buffStreamer == nil {
		return nil
	}
	posSec := i.buffStreamer.getPositionSeconds()
	if posSec == nil {
		return nil
	}
	return &model.Metadata{
		Title:           i.buffStreamer.getTitle(*posSec),
		PlaybackTimeSec: posSec,
	}
}

func (i *Internal) Seek(amtSec int) *model.Metadata {
	if i.cfg.BufferSeconds > 0 {
		i.buffStreamer.seekSec(amtSec)
		return i.Metadata()
	}
	return nil
}

func (i *Internal) Close() error { return nil }
