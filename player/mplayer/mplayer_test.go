package mplayer

import (
	"context"
	"testing"
)

func TestMplayer(t *testing.T) {
	ctx := context.Background()
	p, err := New(ctx, 100)
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
	url = "http://89.238.227.6:8006/;"

	err = p.Play(url)
	if err != nil {
		t.Fatal(err)
	}

	m := p.Metadata()
	if m != nil && m.Err != nil {
		t.Fatal(m.Err)
	}

	err = p.Pause(true)
	if err != nil {
		t.Fatal(err)
	}

	err = p.Pause(false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.SetVolume(50)
	if err != nil {
		t.Fatal(err)
	}

	err = p.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
