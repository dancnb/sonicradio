package player

type Player interface {
	Init() error
	Play(url string) error
	Pause() error
	Stop() error
	Metadata() *Metadata
	Quit() error
}

type Metadata struct {
	Title string
	URL   string
	Err   error
}
