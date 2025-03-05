package vlc

import (
	"context"
	"testing"
)

func TestVlc(t *testing.T) {
	ctx := context.Background()
	p, err := NewVlc(ctx)
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
	m = p.Seek(-15)
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
	_, err = p.SetVolume(20)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.SetVolume(100)
	if err != nil {
		t.Fatal(err)
	}

	err = p.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
