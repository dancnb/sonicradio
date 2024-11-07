package player

import "io"

type Player interface {
	io.Closer

	Play(url string) error
	Pause(value bool) error
	Stop() error
	SetVolume(value int) error
	Metadata() *Metadata
}

type Metadata struct {
	Title string
	Err   error
}
