package player

type Player interface {
	Play(url string) error
	Pause(value bool) error
	Stop() error
	SetVolume(value int) error
	Metadata() *Metadata
	Close() error
}

type Metadata struct {
	Title        string
	PlaybackTime *float64
	Err          error
}
