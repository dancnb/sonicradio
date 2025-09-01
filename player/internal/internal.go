package internal

import (
	"context"

	"github.com/dancnb/sonicradio/player/model"
	"github.com/gopxl/beep/v2/speaker"
)

type Internal struct {

	// streamer
	cancelFn     context.CancelFunc
	buffStreamer *bufferedStreamer
}

func New(ctx context.Context) *Internal { return &Internal{} }

func (b *Internal) Play(url string) error {
	b.Stop()

	var ctx context.Context
	ctx, b.cancelFn = context.WithCancel(context.Background())
	var err error
	b.buffStreamer, err = playStream(ctx, url)
	return err
}

func (b *Internal) Pause(value bool) error {
	speaker.Lock()
	b.buffStreamer.ctrl.Paused = !b.buffStreamer.ctrl.Paused
	speaker.Unlock()
	return nil
}

func (b *Internal) Stop() error {
	if b.cancelFn != nil {
		b.cancelFn()
	}
	return nil
}

// TODO
func (b *Internal) SetVolume(value int) (int, error) {
	// speaker.Lock()
	// b.buffStreamer.volume.Volume -= 0.1
	// speaker.Unlock()
	return -1, nil
}

// TODO
func (b *Internal) Metadata() *model.Metadata { return nil }

// TODO
func (b *Internal) Seek(amtSec int) *model.Metadata { return nil }

// TODO
func (b *Internal) Close() error { return nil }
