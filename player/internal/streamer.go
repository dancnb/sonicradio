package internal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/vorbis"
)

const (
	defNetworkChunkSize = 1024
	defBufferChunkSize  = 1024
	defBufferSize       = 16384
)

type bufferedStreamer struct {
	samples chan [2]float64
	closed  bool
}

func newBufferedStreamer(bufSize int) *bufferedStreamer {
	return &bufferedStreamer{
		samples: make(chan [2]float64, bufSize),
	}
}

func (s *bufferedStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	for i := range samples {
		val, more := <-s.samples
		if !more {
			return i, i > 0
		}
		samples[i] = val
	}
	return len(samples), true
}

func (s *bufferedStreamer) Err() error { return nil }

func (s *bufferedStreamer) Close() {
	if !s.closed {
		close(s.samples)
		s.closed = true
	}
}

func playStream(ctx context.Context, url string) error {
	// -- Network read
	resp, metaInfo, err := openStream(ctx, url)
	if err != nil {
		return fmt.Errorf("open stream err: %w", err)
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	audioPipeR, audioPipeW := io.Pipe()
	go readStream(ctx, audioPipeW, reader, int64(metaInfo.Metaint))

	// -- Decode
	decoderFn, err := getDecoder(metaInfo.ContentType)
	if err != nil {
		return fmt.Errorf("get decoder err: %w", err)
	}
	beepStreamer, format, err := decoderFn(audioPipeR)
	if err != nil {
		return fmt.Errorf("call decoder err: %w", err)
	}
	defer beepStreamer.Close()

	// -- Buffer
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	bufStreamer := newBufferedStreamer(defBufferSize)
	defer bufStreamer.Close()
	go bufferStream(ctx, bufStreamer, beepStreamer)

	// -- Play
	ctrl := &beep.Ctrl{Streamer: bufStreamer, Paused: false}
	resampler := beep.ResampleRatio(4, 1, ctrl)
	volume := &effects.Volume{
		Streamer: resampler,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}
	speaker.Play(volume)

	select {} // keep running
}

// metaInfo contains icy headers.  Example(https://icecast.walmradio.com:8443/otr_opus):
//
//	"ice-audio-info: ice-bitrate=64;ice-channels=1;ice-samplerate=48000"
//	"icy-pub: 0"
//	"icy-index-metaInfo: 1"
//	"icy-logo: https://icecast.walmradio.com:8443/otr.jpg"
//	"icy-country-code: US"
//	"icy-country-subdivision-code: US-NY"
//	"icy-language-codes: en"
//	"icy-main-stream-url: https://icecast.walmradio.com:8443/otr_opus"
//	"icy-geo-lat-long: 40.75166,-73.97538"
//	"icy-br: 32"
//	"icy-genre: OTR,Old Time Radio,Vintage,Classic,V-Disc,WALM,78,78-RPM,78RPM,Easy Listening,Comedy,Drama,Mystery,Sci-Fi,Musical,WALM,Opus"
//	"icy-name: WALM - Old Time Radio Opus"
//	"icy-description: The Golden Age of Radio"
//	"icy-url: https://walmradio.com/otr"
type metaInfo struct {
	// Name string
	// Br int
	// Notice1 string
	// URL string
	// Notice2 string
	// Genre string
	// Pub int

	Metaint     int
	Sr          int
	ContentType string
}

func openStream(ctx context.Context, url string) (resp *http.Response, metaInfo metaInfo, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Icy-MetaData", "1")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if val := resp.Header.Get("icy-metaint"); val != "" {
		fmt.Sscanf(val, "%d", &metaInfo.Metaint)
	}
	if val := resp.Header.Get("icy-sr"); val != "" {
		fmt.Sscanf(val, "%d", &metaInfo.Sr)
	} else if val := resp.Header.Get("ice-audio-info"); val != "" {
		fields := strings.Split(val, ";")
		for _, f := range fields {
			if strings.Contains(f, "samplerate") {
				if p := strings.Split(f, "="); len(p) == 2 {
					fmt.Sscanf(p[1], "%d", &metaInfo.Sr)
					break
				}
			}
		}
	}
	if val := resp.Header.Get("content-type"); val != "" {
		metaInfo.ContentType = val
	}
	return
}

func readStream(
	ctx context.Context,
	wc io.WriteCloser,
	r *bufio.Reader,
	metaInt int64,
) {
	log := slog.With("caller", "readStream")
	defer wc.Close()

	chunkByteSize := metaInt
	if chunkByteSize == 0 {
		chunkByteSize = defNetworkChunkSize
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("read stream cancelled")
			return

		default:
			n, err := io.CopyN(wc, r, chunkByteSize)
			if err != nil {
				log.Error(fmt.Sprintf("read from stream audio data err: %v", err.Error()))
				return
			}
			log.Info(fmt.Sprintf("copied %d bytes to audio pipe", n))
			if metaInt == 0 {
				continue
			}

			metaLenByte, err := r.ReadByte()
			if err != nil {
				log.Error(fmt.Sprintf("read from stream metadata length err: %v", err.Error()))
				return
			}
			metaLen := int(metaLenByte) * 16
			if metaLen > 0 {
				metaData := make([]byte, metaLen)
				n, err := io.ReadFull(r, metaData)
				if err != nil {
					log.Error(fmt.Sprintf("read from stream metadata content err: %v", err.Error()))
					return
				}
				log.Info(fmt.Sprintf("read %d metdata bytes", n))

				metaStr := string(metaData)
				log.Info("--- metadata: " + metaStr)
				if strings.Contains(metaStr, "StreamTitle='") {
					start := strings.Index(metaStr, "StreamTitle='") + len("StreamTitle='")
					end := strings.Index(metaStr[start:], "';")
					if end > 0 {
						title := metaStr[start : start+end]
						log.Info(" ----------------------->>   Now playing " + title)
					}
				}
			}
		}
	}
}

func bufferStream(ctx context.Context, bufStreamer *bufferedStreamer, streamer beep.StreamSeekCloser) {
	log := slog.With("caller", "bufferStream")
	samples := make([][2]float64, defBufferChunkSize)
	for {
		select {
		case <-ctx.Done():
			return

		default:
			n, ok := streamer.Stream(samples)
			log.Info(fmt.Sprintf("streamed %d samples from beep streamer", n))
			if !ok {
				break
			}
			for i := 0; i < n; i++ {
				bufStreamer.samples <- samples[i]
				// log.Info(fmt.Sprintf("sent sample %d to buffered streamer", i))
			}
		}
	}
}

var (
	errAACNotAvailable = errors.New("AAC streams are not supported")
)

func getDecoder(contentType string) (
	func(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error),
	error,
) {
	switch contentType {
	case "audio/mpeg", "audio/x-scpls":
		return mp3.Decode, nil
	case "audio/ogg":
		return vorbis.Decode, nil
	case "audio/aac":
		return nil, errAACNotAvailable
	// TODO: wav?

	default:
		return nil, fmt.Errorf("Stream content-type not supported: %s", contentType)
	}
}
