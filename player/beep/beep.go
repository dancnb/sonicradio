package beep

import (
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player/model"
)

type Beep struct{}

func NewBeep() *Beep { return &Beep{} }

func (b *Beep) GetType() config.PlayerType       { return config.BEEP }
func (b *Beep) Play(url string) error            { return nil }
func (b *Beep) Pause(value bool) error           { return nil }
func (b *Beep) Stop() error                      { return nil }
func (b *Beep) SetVolume(value int) (int, error) { return -1, nil }
func (b *Beep) Metadata() *model.Metadata        { return nil }
func (b *Beep) Seek(amtSec int) *model.Metadata  { return nil }
func (b *Beep) Close() error                     { return nil }
