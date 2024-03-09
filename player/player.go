package player

type Player interface {
	Play(url string) error
	Stop() error
}
