package internal

import (
	"context"
	"log/slog"
	"time"

	"github.com/dancnb/sonicradio/player/model"
	"github.com/gopxl/beep/v2/speaker"
)

type Internal struct {
	// streamer
	cancelFn     context.CancelFunc
	buffStreamer *bufferedStreamer
}

func New(ctx context.Context) *Internal {
	return &Internal{}
}

func (i *Internal) Play(url string) error {
	i.Stop()

	var ctx context.Context
	ctx, i.cancelFn = context.WithCancel(context.Background())
	var err error
	i.buffStreamer, err = newBufferedStreamer(ctx, url)
	if err != nil {
		return err
	}
	return nil
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

func (i *Internal) Metadata() *model.Metadata {
	if i.buffStreamer != nil {
		m := model.Metadata{
			Title: i.buffStreamer.title,
		}
		speaker.Lock()
		pos := i.buffStreamer.beepStreamer.Position()
		posD := i.buffStreamer.format.SampleRate.D(pos)
		posSec := int64(posD.Round(time.Second).Seconds())
		slog.Info("", "pos", pos, "posD", posD)
		m.PlaybackTimeSec = &posSec
		speaker.Unlock()
		return &m
	}
	return nil
}

// TODO
func (i *Internal) Seek(amtSec int) *model.Metadata { return nil }

func (i *Internal) Close() error { return nil }
