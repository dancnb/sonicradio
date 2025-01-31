package ffplay

import (
	"context"
	"testing"
	"time"
)

func TestFFPlay(t *testing.T) {
	ctx := context.Background()
	p, err := NewFFPlay(ctx)
	if err != nil {
		t.Fatal(err)
	}
	p.SetVolume(100)
	defer func() {
		err := p.Stop()
		if err != nil {
			t.Fatal(err)
		}
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

	time.Sleep(200 * time.Millisecond)
	m := p.Metadata()
	if m.Err != nil {
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
	_, err = p.SetVolume(70)
	if err != nil {
		t.Fatal(err)
	}

	err = p.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
