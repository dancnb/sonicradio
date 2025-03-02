package vlc

var baseArgs = []string{"-I", "rc", "--rc-quiet", "--volume-step", "12.8", "--gain", "1.0", "--no-video", "--rc-host"}

const baseCmd = "vlc.exe"

func GetBaseCmd() string {
	return baseCmd
}
