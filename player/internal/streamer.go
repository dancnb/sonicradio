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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/vorbis"
)

const (
	networkReadSize = 4096
	beepReadSize    = 4096

	contentTypePls   = "audio/x-scpls"
	contentTypeMpeg  = "audio/mpeg"
	contentTypeMpeg2 = "audio/x-mpegurl"
	contentTypeOgg   = "audio/ogg"
	contentTypeOgg2  = "application/ogg"
	contentTypeAac   = "audio/aac"
	contentTypeAacp  = "audio/aacp"
)

type bufferedStreamer struct {
	url   string
	title map[int64]string
	wg    sync.WaitGroup

	ch   chan [2]float64
	data [][2]float64
	//
	// write index: where next decoded sample will be written
	wx int64

	rbSync sync.Mutex
	// read back offset relative to write index;
	// values < 0 means it has remaining data from buffer to play
	rbx       int
	streamPos int64
	done      chan struct{}

	beepStreamer beep.StreamSeekCloser // used for getPositionSeconds
	format       beep.Format           // used for getPositionSeconds
	ctrl         *beep.Ctrl            // used for togglePause
	volume       *effects.Volume
}

func newBufferedStreamer(
	ctx context.Context,
	url string,
	volume int,
	buffer [][2]float64,
) (*bufferedStreamer, error) {
	log := slog.With("caller", "newBufferedStreamer", "url", url)
	log.Info("start")
	defer func() { log.Info("end") }()

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
		return newBufferedStreamer(ctx, plsUrl, volume, buffer)
	}

	bs := &bufferedStreamer{
		url:   url,
		title: make(map[int64]string),
		ch:    make(chan [2]float64),
		done:  make(chan struct{}),
		data:  buffer,
	}

	audioPipeR, audioPipeW := io.Pipe()

	titleCh := make(chan string, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-titleCh:
				bs.title[bs.streamPos] = t
			}
		}
	}()

	bs.wg.Add(1)
	go readStream(ctx, &bs.wg, url, audioPipeW, resp.Body, int64(metaInfo.Metaint), titleCh)

	// -- Decode
	// beep.Decode takes a ReadCloser containing audio data in MP3 format and returns a StreamSeekCloser,
	// which streams that audio. The Seek method will panic if rc is not io.Seeker.
	//
	// Do not close the supplied ReadSeekCloser, instead, use the Close method of the returned
	// StreamSeekCloser when you want to release the resources.
	decoderFn, err := getDecoder(metaInfo.ContentType)
	if err != nil {
		return nil, err
	}
	bs.beepStreamer, bs.format, err = decoderFn(audioPipeR)
	if err != nil {
		return nil, err
	}
	slog.Info("", "sampleRate", bs.format.SampleRate)

	bs.wg.Add(1)
	go func() {
		<-ctx.Done()
		bs.Close()
		bs.wg.Done()
		log.Info("===  CANCEL 1 (bufferedStreamer closed) ===")
	}()

	speaker.Init(bs.format.SampleRate, bs.format.SampleRate.N(time.Second/10))

	// -- Buffer
	bs.wg.Add(1)
	go bs.readDecodedSamples(ctx)

	// -- Play
	bs.ctrl = &beep.Ctrl{Streamer: bs, Paused: false}
	expVolume := percentToExponent(float64(volume))
	bs.volume = &effects.Volume{
		Streamer: bs.ctrl,
		Base:     2,
		Volume:   expVolume,
		Silent:   false,
	}
	speaker.Play(bs.volume)

	return bs, nil
}

func (bs *bufferedStreamer) readDecodedSamples(ctx context.Context) {
	log := slog.With("method", "readDecodedSamples", "url", bs.url)
	defer func() {
		close(bs.ch)
		bs.wg.Done()
		log.Info("===  CANCEL 2 (bs.wCh closed) ===")
	}()

	decodedSamples := make([][2]float64, beepReadSize)
	for {
		n, more := bs.beepStreamer.Stream(decodedSamples)
		if !more {
			log.Info("===  CANCEL 2.1 (no more samples in beepStreamer) ===")
			if err := bs.beepStreamer.Err(); err != nil {
				log.Info(fmt.Sprintf("beepStreamer error: %#v", err))
			}
			return
		}
		for i := range n {
			select {
			case <-ctx.Done():
				log.Info("ctx done")
				return
			case bs.ch <- decodedSamples[i]:
				if len(bs.data) > 0 {
					wIdx := bs.wx % int64(len(bs.data))
					bs.data[wIdx] = decodedSamples[i]
					bs.wx++
				}
			}
		}
	}
}

// Stream: while Ctrl is paused, this call is not reached
func (bs *bufferedStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	bs.rbSync.Lock()
	defer bs.rbSync.Unlock()

	log := slog.With("method", "Stream", "url", bs.url)

	i := 0

	if len(bs.data) > 0 {
		// first check for remaining buffered samples
		for bs.rbx < 0 && i < len(samples) {
			select {
			case <-bs.done:
				log.Info("===  CANCEL 3.2 (bs.done) ===")
				return 0, false
			default:
				n := int64(len(bs.data))
				idx := (bs.wx + int64(bs.rbx) + n) % n
				bs.rbx++
				//skip empty buffer data
				if idx >= bs.wx {
					continue
				}
				val := bs.data[idx]
				samples[i] = val
				i++
			}
		}
		// filled samples completely from buffered data
		if i == len(samples) {
			return i, true
		}
	}

	// fill remaining from decoded channel
	for i < len(samples) {
		val, more := <-bs.ch
		if !more {
			return i, i > 0
		}
		samples[i] = val
		i++
	}

	return len(samples), len(samples) > 0
}

func (bs *bufferedStreamer) seekSec(amtSec int) {
	if len(bs.data) == 0 {
		return
	}

	bs.rbSync.Lock()
	defer bs.rbSync.Unlock()

	log := slog.With("method", "bufferedStreamer.seekSec", "url", bs.url)

	pos, delta := bs.rbx, 0
	log.Info("", "currPos", pos)
	if amtSec > 0 {
		delta = bs.secondsToSamples(amtSec)
	} else {
		delta = -bs.secondsToSamples(-amtSec)
	}
	pos += delta
	log.Info("", "newPos unclamped", pos)
	// clamp the position
	pos = max(pos, -len(bs.data))
	pos = min(pos, 0)
	log.Info("", "newPos clamped", pos)

	bs.rbx = pos
}

func (bs *bufferedStreamer) secondsToSamples(sec int) int {
	return bs.format.SampleRate.N(time.Second * time.Duration(sec))
}

func (bs *bufferedStreamer) samplesToSeconds(s int) int {
	return int(bs.format.SampleRate.D(s).Round(time.Second).Seconds())
}

func (bs *bufferedStreamer) togglePause() {
	if bs == nil {
		return
	}
	speaker.Lock()
	bs.ctrl.Paused = !bs.ctrl.Paused
	speaker.Unlock()
}

func (bs *bufferedStreamer) getTitle(posSec int64) string {
	ts := make([]int64, len(bs.title))
	i := 0
	for k := range bs.title {
		ts[i] = k
		i++
	}
	slices.Sort(ts)
	for i := len(ts) - 1; i >= 0; i-- {
		if ts[i] > posSec {
			continue
		}
		return bs.title[ts[i]]
	}
	return ""
}

func (bs *bufferedStreamer) getPositionSeconds() *int64 {
	if bs == nil {
		return nil
	}

	// bs.rbSync.Lock()
	// defer bs.rbSync.Unlock()

	if bs.rbx == 0 {
		bs.streamPos = bs.getStreamPosition()
		return &bs.streamPos
	}

	backSec := bs.samplesToSeconds(-bs.rbx)
	backPos := bs.streamPos - int64(backSec)
	backPos = max(0, backPos)
	slog.Info("", "stream position", bs.streamPos, "back position", backPos)
	return &backPos
}

func (bs *bufferedStreamer) getStreamPosition() int64 {
	speaker.Lock()
	pos := bs.beepStreamer.Position()
	posD := bs.format.SampleRate.D(pos)
	speaker.Unlock()
	posSec := int64(posD.Round(time.Second).Seconds())
	slog.Info("", "stream position", posSec)
	return posSec
}

func (bs *bufferedStreamer) setVolumeFromPercentage(value int) {
	if bs == nil {
		return
	}
	speaker.Lock()
	expValue := percentToExponent(float64(value))
	bs.volume.Volume = expValue
	speaker.Unlock()
}

func (bs *bufferedStreamer) Err() error { return nil }

func (bs *bufferedStreamer) Close() error {
	close(bs.done)

	if bs.beepStreamer == nil {
		return nil
	}
	if err := bs.beepStreamer.Close(); err != nil {
		return fmt.Errorf("beepStreamer close err: %w", err)
	}
	return nil
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
	log := slog.With("caller", "openStream")
	log.Info("start")
	defer func() {
		log.Info("end")
		if err != nil {
			log.Info(fmt.Sprintf("end err=%v", err))
		}
	}()

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
	wg *sync.WaitGroup,
	url string,
	wc io.WriteCloser,
	respBody io.ReadCloser,
	metaInt int64,
	titleCh chan string,
) {
	log := slog.With("caller", "readStream", "url", url)
	log.Info("start")
	defer func() {
		err := respBody.Close()
		log.Info(fmt.Sprintf("http response body close err: %#v", err))
		err = wc.Close()
		log.Info(fmt.Sprintf("audio pipe writer body close err: %#v", err))
		close(titleCh)
		wg.Done()
		log.Info("end")
	}()

	chunkByteSize := metaInt
	if chunkByteSize == 0 {
		chunkByteSize = networkReadSize
	}
	bufReader := bufio.NewReader(respBody)
	for {
		select {
		case <-ctx.Done():
			log.Info("ctx cancelled")
			return

		default:
			_, err := io.CopyN(wc, bufReader, chunkByteSize)
			if err != nil {
				log.Error(fmt.Sprintf("read from stream audio data err: %v", err.Error()))
				return
			}
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
	case contentTypeMpeg, contentTypeMpeg2:
		return mp3.Decode, nil
	case contentTypeOgg, contentTypeOgg2:
		return vorbis.Decode, nil
	case contentTypeAac, contentTypeAacp:
		return nil, errAACNotAvailable
	// TODO: wav?

	default:
		return nil, fmt.Errorf("Stream content-type not supported: %s", contentType)
	}
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
