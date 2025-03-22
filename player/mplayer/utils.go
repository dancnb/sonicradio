//go:build !windows

package mplayer

const baseCmd = "mplayer"

func GetBaseCmd() string {
	return baseCmd
}
