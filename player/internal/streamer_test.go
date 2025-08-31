package internal

import (
	"context"
	"testing"
)

func Test_openStream(t *testing.T) {
	url := "http://vibration.stream2net.eu:8220/;stream/1"
	// url = "https://icecast.walmradio.com:8443/otr_opus"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, info, err := openStream(ctx, url)
	if err != nil {
		t.Errorf("open stream err=%v", err)
	} else if resp == nil {
		t.Error("no stream response")
	}
	t.Logf("%#v", info)
}

func Test_playStream(t *testing.T) {
	url := "http://vibration.stream2net.eu:8220/;stream/1"
	// url = "https://icecast.walmradio.com:8443/otr_opus"
	// url = "http://play.strefa.fm:8000/strefa.ogg"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	playStream(ctx, url)
}
