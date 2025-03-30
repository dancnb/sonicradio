//go:build !windows

package mpd

const baseCmd = "mpd"

func GetBaseCmd() string {
	return baseCmd
}
