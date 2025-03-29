//go:build !windows

package mpd

const baseCmd = "mpc"

func GetBaseCmd() string {
	return baseCmd
}
