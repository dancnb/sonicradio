package internal

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
)

// http://vibration.stream2net.eu:8220/;stream/1
// internal.metaInfo{Metaint:16384, Sr:44100, ContentType:"audio/mpeg"}
//
// http://oceanwaves.radio.mynoise.net/
// internal.metaInfo{Metaint:16000, Sr:44100, ContentType:"audio/mpeg"}
//
// https://icecast.walmradio.com:8443/otr_opus
// internal.metaInfo{Metaint:0, Sr:48000, ContentType:"audio/ogg"}
//
// http://radiocdn.nxthost.com/radio-deea
// internal.metaInfo{Metaint:16000, Sr:0, ContentType:"audio/mpeg"}
//
// http://cast.streams.ovh:8008/stream
// internal.metaInfo{Metaint:8192, Sr:44100, ContentType:"audio/mpeg"}
//
// https://icecast.walmradio.com:8443/walm
// internal.metaInfo{Metaint:16000, Sr:48000, ContentType:"audio/mpeg"}
func Test_openStream(t *testing.T) {
	urls := []string{
		"http://vibration.stream2net.eu:8220/;stream/1",
		"http://oceanwaves.radio.mynoise.net/",
		"https://icecast.walmradio.com:8443/otr_opus",
		"http://radiocdn.nxthost.com/radio-deea",
		// "https://cast.streams.ovh:2199/tunein/tranceathena.pls",
		"http://cast.streams.ovh:8008/stream",
		"https://icecast.walmradio.com:8443/walm",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, url := range urls {
		slog.Info(url)
		resp, info, err := openStream(ctx, url)
		if err != nil {
			t.Errorf("open stream err=%v", err)
		} else if resp == nil {
			t.Error("no stream response")
		}
		slog.Info(fmt.Sprintf("%#v", info))
	}
}

func Test_playStream(t *testing.T) {
	url := "http://vibration.stream2net.eu:8220/;stream/1"
	// url = "http://oceanwaves.radio.mynoise.net/"
	// url = "https://icecast.walmradio.com:8443/otr_opus"
	// url = "http://play.strefa.fm:8000/strefa.ogg"
	// url = "http://cast.streams.ovh:8008/stream"
	url = "https://icecast.walmradio.com:8443/walm"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	playStream(ctx, url)
}

//TODO pls file:
// [playlist]
// numberofentries=1
// File1=http://cast.streams.ovh:8008/stream
// Title1=Trance Athena Radio
// Length1=-1
// version=2
