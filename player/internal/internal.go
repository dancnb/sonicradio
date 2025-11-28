package internal

import (
	"context"
	"log/slog"
	"time"

	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
	"github.com/gopxl/beep/v2"
)

const defSampleRate = 44100

type Internal struct {
	volume int
	cfg    config.InternalPlayer
	buffer [][2]float64

	// streamer
	cancelFn     context.CancelFunc
	buffStreamer *bufferedStreamer
}

func New(ctx context.Context, volume int, cfg config.InternalPlayer) *Internal {
	return &Internal{
		volume: volume,
		cfg:    cfg,
		buffer: newBuffer(cfg.BufferSeconds),
	}
}

func (i *Internal) Play(url string) error {
	log := slog.With("caller", "Internal.Play", "url", url)
	log.Info("start")
	defer func() { log.Info("end") }()

	_ = i.Stop()

	var ctx context.Context
	ctx, cancelFn := context.WithCancel(context.Background())
	clear(i.buffer)
	buffStreamer, err := newBufferedStreamer(ctx, url, i.volume, i.buffer)
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

func newBuffer(bufferSeconds int) [][2]float64 {
	if bufferSeconds <= 0 {
		return nil
	}
	sr := beep.SampleRate(defSampleRate)
	buffLen := sr.N(time.Duration(bufferSeconds) * time.Second)
	slog.Info("newBuffer",
		"bufferSeconds", bufferSeconds,
		"buffLen", buffLen,
		"size", float64(buffLen*2*8)/1000000,
	)
	return make([][2]float64, buffLen)
}
