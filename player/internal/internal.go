package internal

import (
	"context"

	"github.com/dancnb/sonicradio/player/model"
)

type Internal struct {
	volume int

	// streamer
	cancelFn     context.CancelFunc
	buffStreamer *bufferedStreamer
}

func New(ctx context.Context, volume int) *Internal {
	return &Internal{volume: volume}
}

func (i *Internal) Play(url string) error {
	i.Stop()

	var ctx context.Context
	ctx, i.cancelFn = context.WithCancel(context.Background())
	var err error
	i.buffStreamer, err = newBufferedStreamer(ctx, url, i.volume)
	if err != nil {
		return err
	}
	return nil
}

func (i *Internal) Pause(value bool) error {
	i.buffStreamer.togglePause()
	return nil
}

func (i *Internal) Stop() error {
	if i.cancelFn != nil {
		i.cancelFn()
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
	return &model.Metadata{
		Title:           i.buffStreamer.title,
		PlaybackTimeSec: i.buffStreamer.getPositionSeconds(),
	}
}

// TODO
func (i *Internal) Seek(amtSec int) *model.Metadata { return nil }

func (i *Internal) Close() error { return nil }
