//go:build !windows

package vlc

// var baseArgs = []string{"-I rc", "--no-video", "--volume-step 12.8", "--gain 1.0"}
// var baseArgs = []string{"-I", "rc", "--rc-fake-tty", "--volume-step", "12.8", "--gain", "1.0", "--no-video", "--rc-host"}
var baseArgs = []string{"-I", "rc", "--volume-step", "12.8", "--gain", "1.0", "--no-video", "--rc-host"}

const baseCmd = "vlc"

func GetBaseCmd() string {
	return baseCmd
}
