package components

import (
	"log/slog"
	"time"
)

const jumpTimeout = 250 * time.Millisecond

type JumpInfo struct {
	position int
	last     time.Time
}

func (j *JumpInfo) JumpTimeout() time.Duration {
	return jumpTimeout
}

func (j *JumpInfo) isActive() bool {
	return j.last.Add(jumpTimeout).After(time.Now())
}

func (j *JumpInfo) NewPosition(digit int) int {
	log := slog.With("method", "components.JumpInfo.getJumpIdx")
	log.Debug("", "digit", digit, "oldPos", j.position)
	if j.isActive() {
		j.position = j.position*10 + digit
	} else {
		j.position = digit
	}
	j.last = time.Now()
	log.Debug("", "newPos", j.position)
	return j.position
}

func (j *JumpInfo) LastPosition() int {
	return j.position
}
