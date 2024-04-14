package player

type Player interface {
	Play(url string) error
	Stop() error
	Metadata() *Metadata
}

type Metadata struct {
	Title string
	URL   string
	Err   error
}
