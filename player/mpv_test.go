package player

import "testing"

func Test_mpv(t *testing.T) {
	p := NewMPV()
	url := "http://radiocdn.nxthost.com/radio-deea"
	err := p.Play(url)
	if err != nil {
		t.Error(err)
	}

	err = p.Stop()
	if err != nil {
		t.Error(err)
	}
}
