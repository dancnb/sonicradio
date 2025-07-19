package mpd

import (
	"context"
	"testing"

	"github.com/dancnb/sonicradio/config"
)

func TestMplayer(t *testing.T) {
	ctx := context.Background()
	p, err := New(ctx, config.DefMpdHost, config.DefMpdPort, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = p.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	streamUrl := "http://stream-uk1.radioparadise.com/aac-320"
	// streamUrl = "http://89.238.227.6:8006/;"
	// streamUrl = "http://radiocdn.nxthost.com/radio-deea"
	streamUrl = "http://vibration.stream2net.eu:8220/;stream/1"

	err = p.Play(streamUrl)
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

	m = p.Seek(-5)
	if m != nil && m.Err != nil {
		t.Fatal(m.Err)
	}

	err = p.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
