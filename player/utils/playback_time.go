package playerutils

import (
	"sync"
	"time"
)

type PlaybackTime struct {
	playTimeMtx   sync.Mutex
	playedTime    *time.Duration
	playStartTime time.Time
}

func NewPlaybackTime() *PlaybackTime {
	return &PlaybackTime{}
}

func (f *PlaybackTime) ResetPlayTime() {
	f.playTimeMtx.Lock()
	defer f.playTimeMtx.Unlock()

	f.playedTime = nil
	f.playStartTime = time.Now()
}

func (f *PlaybackTime) PausePlayTime() {
	f.playTimeMtx.Lock()
	defer f.playTimeMtx.Unlock()

	d := time.Since(f.playStartTime)
	if f.playedTime == nil {
		f.playedTime = &d
	} else {
		*f.playedTime += d
	}
	f.playStartTime = time.Time{}
}

func (f *PlaybackTime) ResumePlayTime() {
	f.playTimeMtx.Lock()
	defer f.playTimeMtx.Unlock()

	f.playStartTime = time.Now()
}

func (f *PlaybackTime) GetPlayTime() *int64 {
	f.playTimeMtx.Lock()
	defer f.playTimeMtx.Unlock()

	var d time.Duration
	if f.playedTime != nil {
		d += *f.playedTime
	}
	if !f.playStartTime.IsZero() {
		d += time.Since(f.playStartTime)
	}
	intD := int64(d.Seconds())
	return &intD
}
