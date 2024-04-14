package player

import "testing"

func Test_mpv(t *testing.T) {
	p := NewMPV()
	url := "http://stream-uk1.radioparadise.com/aac-320"
	// url := "https://dancewaveee.com"
	err := p.Play(url)
	if err != nil {
		t.Error(err)
	}

	err = p.Stop()
	if err != nil {
		t.Error(err)
	}
}
