//go:build !windows

package ffplay

const baseCmd = "ffplay"

func GetBaseCmd() string {
	return baseCmd
}
