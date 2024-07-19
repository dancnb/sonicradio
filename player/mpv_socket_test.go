package player

import (
	"testing"
)

func TestMpvSocket_Play(t *testing.T) {
	p := NewMPVSocket()
	err := p.Init()
	if err != nil {
		t.Fatal(err)
	}
	url := "http://stream-uk1.radioparadise.com/aac-320"
	err = p.Play(url)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Stop()
	if err != nil {
		t.Fatal(err)
	}
	err = p.Quit()
	if err != nil {
		t.Fatal(err)
	}
}
