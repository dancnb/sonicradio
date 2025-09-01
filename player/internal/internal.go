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

func (i *Internal) Play(url string) error {
	i.Stop()

	var ctx context.Context
	ctx, i.cancelFn = context.WithCancel(context.Background())
	var err error
	i.buffStreamer, err = playStream(ctx, url)
	return err
}

func (i *Internal) Pause(value bool) error {
	speaker.Lock()
	i.buffStreamer.ctrl.Paused = !i.buffStreamer.ctrl.Paused
	speaker.Unlock()
	return nil
}

func (i *Internal) Stop() error {
	if i.cancelFn != nil {
		i.cancelFn()
	}
	return nil
}

// TODO
func (i *Internal) SetVolume(value int) (int, error) {
	// speaker.Lock()
	// b.buffStreamer.volume.Volume -= 0.1
	// speaker.Unlock()
	return -1, nil
}

// TODO
func (i *Internal) Metadata() *model.Metadata { return nil }

// TODO
func (i *Internal) Seek(amtSec int) *model.Metadata { return nil }

func (i *Internal) Close() error { return nil }
