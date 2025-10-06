package internal

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
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
	defNetworkChunkSize = 4096
	defBufferChunkSize  = 4096
	defBufferSize       = 16384

	contentTypePls  = "audio/x-scpls"
	contentTypeMpeg = "audio/mpeg"
	contentTypeOgg  = "audio/ogg"
	contentTypeOgg2 = "application/ogg"
	contentTypeAac  = "audio/aac"
	contentTypeAacp = "audio/aacp"
)

type bufferedStreamer struct {
	url   string
	title string

	samples chan [2]float64
	closed  bool

	beepStreamer beep.StreamSeekCloser
	format       beep.Format
	ctrl         *beep.Ctrl
	resampler    *beep.Resampler
	volume       *effects.Volume
}

func newBufferedStreamer(ctx context.Context, url string, volume int) (*bufferedStreamer, error) {
	// -- Network read
	resp, metaInfo, err := openStream(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("open stream err: %w", err)
	}

	if strings.ToLower(metaInfo.ContentType) == contentTypePls {
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read pls file: %w", err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(b))
		var plsUrl string
		for scanner.Scan() {
			l := scanner.Text()
			if strings.Contains(strings.ToLower(l), "file1") {
				if p := strings.Split(l, "="); len(p) == 2 {
					plsUrl = strings.TrimSpace(p[1])
					break
				}
			}
		}
		if plsUrl == "" {
			return nil, fmt.Errorf("could not parse URL from playlist file [%s]", url)
		}
		return newBufferedStreamer(ctx, plsUrl, volume)
	}

	audioPipeR, audioPipeW := io.Pipe()
	titleCh := make(chan string, 1)
	go readStream(ctx, url, audioPipeW, resp.Body, int64(metaInfo.Metaint), titleCh)

	// -- Decode
	decoderFn, err := getDecoder(metaInfo.ContentType)
	if err != nil {
		return nil, err
	}
	beepStreamer, format, err := decoderFn(audioPipeR)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		slog.With("caller", "newBufferedStreamer", "url", url).
			Info("===  CANCEL 1 (beepStreamer close) ===")
		beepStreamer.Close()
	}()
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// -- Buffer
	bufStreamer := &bufferedStreamer{
		url:          url,
		samples:      make(chan [2]float64, defBufferSize),
		beepStreamer: beepStreamer,
		format:       format,
	}
	go bufStreamer.doBuffer(beepStreamer)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-titleCh:
				bufStreamer.title = t
			}
		}
	}()

	// -- Play
	bufStreamer.ctrl = &beep.Ctrl{Streamer: bufStreamer, Paused: false}
	bufStreamer.resampler = beep.ResampleRatio(4, 1, bufStreamer.ctrl)
	expVolume := percentToExponent(float64(volume))
	bufStreamer.volume = &effects.Volume{
		Streamer: bufStreamer.resampler,
		Base:     2,
		Volume:   expVolume,
		Silent:   false,
	}
	speaker.Play(bufStreamer.volume)

	return bufStreamer, nil
}

func (bs *bufferedStreamer) doBuffer(beepStreamer beep.StreamSeekCloser) {
	log := slog.With("method", "bufferedStreamer.doBuffer", "url", bs.url)
	defer func() {
		log.Info("===  CANCEL 3 (bufStreamer close) ===")
		bs.Close()
	}()

	samples := make([][2]float64, defBufferChunkSize)
	for {
		n, more := beepStreamer.Stream(samples)
		log.Info(fmt.Sprintf("beepStreamer ---> %d samples, more=%v ---> bufferedStreamer", n, more))
		if !more {
			log.Info("===  CANCEL 2 (no more samples in beepStreamer) ===")
			if err := beepStreamer.Err(); err != nil {
				log.Info(fmt.Sprintf("beepStreamer error: %#v", err))
			}
			break
		}
		for i := range n {
			bs.samples <- samples[i]
			// log.Info(fmt.Sprintf("beepStreamer -> sample %d -> bufferedStreamer", i))
		}
	}
}

func (bs *bufferedStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	log := slog.With("method", "bufferedStreamer.doBuffer", "url", bs.url)
	for i := range samples {
		val, more := <-bs.samples
		if !more {
			log.Info("===  CANCEL 4 (no more samples in bufferedStreamer) ===")
			return i, i > 0
		}
		samples[i] = val
	}
	return len(samples), true
}

func (bs *bufferedStreamer) togglePause() {
	if bs == nil {
		return
	}
	speaker.Lock()
	bs.ctrl.Paused = !bs.ctrl.Paused
	speaker.Unlock()
}

func (bs *bufferedStreamer) getPositionSeconds() *int64 {
	if bs == nil {
		return nil
	}
	speaker.Lock()
	pos := bs.beepStreamer.Position()
	posD := bs.format.SampleRate.D(pos)
	speaker.Unlock()
	posSec := int64(posD.Round(time.Second).Seconds())
	slog.Info("", "pos", pos, "posD", posD)
	return &posSec
}
func (bs *bufferedStreamer) setVolumeFromPercentage(value int) {
	if bs == nil {
		return
	}
	log := slog.With("method", "bufferedStreamer.setVolumeFromPercentage")
	speaker.Lock()
	expValue := percentToExponent(float64(value))
	bs.volume.Volume = expValue
	speaker.Unlock()
	log.Info("", "perc", value, "exp", expValue)
}

func percentToExponent(p float64) float64 {
	minExp := -10.0
	curve := 0.5
	if p <= 0 {
		return minExp
	}
	if p >= 100 {
		return 0
	}
	n := p / 100.0
	adjusted := math.Pow(n, curve)
	return (1.0 - adjusted) * minExp
}

func (bs *bufferedStreamer) Err() error { return nil }

func (bs *bufferedStreamer) Close() {
	if !bs.closed {
		close(bs.samples)
		bs.closed = true
	}
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

func openStream(
	ctx context.Context,
	url string,
) (
	resp *http.Response,
	metaInfo metaInfo,
	err error,
) {
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
	url string,
	wc io.WriteCloser,
	respBody io.ReadCloser,
	metaInt int64,
	titleCh chan string,
) {
	log := slog.With("caller", "readStream", "url", url)
	bufReader := bufio.NewReader(respBody)

	defer func() {
		err := respBody.Close()
		log.Info(fmt.Sprintf("http response body close err: %#v", err))
		err = wc.Close()
		log.Info(fmt.Sprintf("audio pipe writer body close err: %#v", err))
		close(titleCh)
	}()

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
			n, err := io.CopyN(wc, bufReader, chunkByteSize)
			if err != nil {
				log.Error(fmt.Sprintf("read from stream audio data err: %v", err.Error()))
				return
			}
			log.Info(fmt.Sprintf("network ---> copied %d bytes ---> audio pipe (beepStreamer)", n))
			if metaInt == 0 {
				continue
			}

			metaLenByte, err := bufReader.ReadByte()
			if err != nil {
				log.Error(fmt.Sprintf("read from stream metadata length err: %v", err.Error()))
				return
			}
			metaLen := int(metaLenByte) * 16
			if metaLen > 0 {
				metaData := make([]byte, metaLen)
				n, err := io.ReadFull(bufReader, metaData)
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
						go func() {
							titleCh <- title
						}()
					}
				}
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
	case contentTypeMpeg:
		return mp3.Decode, nil
	case contentTypeOgg,contentTypeOgg2:
		return vorbis.Decode, nil
	case contentTypeAac, contentTypeAacp:
		return nil, errAACNotAvailable
	// TODO: wav?

	default:
		return nil, fmt.Errorf("Stream content-type not supported: %s", contentType)
	}
}
