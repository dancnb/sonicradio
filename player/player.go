package player

type Player interface {
	Play(url string) (string, error)
	Stop() error
}
