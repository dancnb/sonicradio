package mpv

import (
	"context"
	"testing"
)

func TestMpvSocket_Play(t *testing.T) {
	ctx := context.Background()
	p, err := NewMPVSocket(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = p.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	url := "http://stream-uk1.radioparadise.com/aac-320"

	err = p.Play(url)
	if err != nil {
		t.Fatal(err)
	}

	m := p.Metadata()
	if m.Err != nil {
		t.Fatal(m.Err)
	}
	mt := p.getMediaTitle()
	if mt.Err != nil {
		t.Fatal(m.Err)
	}
	m = p.Seek(-5)
	if m.Err != nil {
		t.Fatal(err)
	}

	err = p.Pause(true)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Pause(false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.SetVolume(70)
	if err != nil {
		t.Fatal(err)
	}

	err = p.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
