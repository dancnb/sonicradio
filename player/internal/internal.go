package internal

import (
	"context"

	"github.com/dancnb/sonicradio/player/model"
)

type Internal struct{}

func New(ctx context.Context) *Internal { return &Internal{} }

func (b *Internal) Play(url string) error            { return nil }
func (b *Internal) Pause(value bool) error           { return nil }
func (b *Internal) Stop() error                      { return nil }
func (b *Internal) SetVolume(value int) (int, error) { return -1, nil }
func (b *Internal) Metadata() *model.Metadata        { return nil }
func (b *Internal) Seek(amtSec int) *model.Metadata  { return nil }
func (b *Internal) Close() error                     { return nil }
