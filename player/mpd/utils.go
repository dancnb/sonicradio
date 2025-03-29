//go:build !windows

package mpd

var baseArgs = []string{
	"-w",
}

const baseCmd = "mpc"

func GetBaseCmd() string {
	return baseCmd
}
