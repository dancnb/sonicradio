package player

type Player interface {
	Play(url string) error
	Stop() error
	Metadata() (*Metadata, error)
}

type Metadata struct {
	Title string
}
