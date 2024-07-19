package player

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"runtime"
	"slices"
)

var (
	baseSockArgs = []string{"--idle", "--terminal=no", "--no-video"}
	ipcArg       = "--input-ipc-server=%s"
	sockFile     = "/tmp/mpvsocket.%d"
	sockFileWin  = `\\.\pipe\mpvsocket.%d`
)

type ipcCmd uint8

const (
	play ipcCmd = iota
	stop
	quit
)

var ipcCmds = map[ipcCmd]string{
	play: `{ "command": ["loadfile", "%s","replace"] }`,
	stop: `{ "command": [ "stop"] }`,
	quit: `{ "command": [ "quit"] }`,
}

func NewMPVSocket() Player {
	p := &MpvSocket{}
	return p
}

type MpvSocket struct {
	sockFile string
	url      string
	cmd      *exec.Cmd
}

func (mpv *MpvSocket) Init() error {
	log := slog.With("method", "MpvSocket.Init")
	log.Debug("init")

	mpv.sockFile = mpv.getSockFilePath()
	args := slices.Clone(baseSockArgs)
	args = append(args, fmt.Sprintf(ipcArg, mpv.sockFile))
	cmd := exec.Command(mpv.getBaseCmd(), args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	} else if cmd.Err != nil {
		log.Error("mpv cmd error", "error", cmd.Err.Error())
		return cmd.Err
	}
	err := cmd.Start()
	if err != nil {
		log.Error("mpv cmd start", "error", err)
		return err
	}
	mpv.cmd = cmd
	log.Debug("mpv cmd started", "pid", mpv.cmd.Process.Pid)

	return nil
}

func (mpv *MpvSocket) Pause() error { return nil }

func (mpv *MpvSocket) Play(url string) error {
	log := slog.With("method", "MpvSocket.Play")
	log.Info("playing url=" + url)

	conn, err := net.Dial("unix", mpv.sockFile)
	if err != nil {
		return err
	}
	playCmd := fmt.Sprintf(ipcCmds[play], url) + "\n"
	_, err = conn.Write([]byte(playCmd))
	if err != nil {
		return err
	}

	b := make([]byte, 1024)
	_, err = conn.Read(b)
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("resp=%s", b))

	return nil
}

func (mpv *MpvSocket) Metadata() *Metadata {
	return &Metadata{}
}

func (mpv *MpvSocket) Stop() error {
	log := slog.With("method", "MpvSocket.Stop")
	log.Info("stopping")

	conn, err := net.Dial("unix", mpv.sockFile)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(ipcCmds[stop] + "\n"))
	if err != nil {
		return err
	}

	b := make([]byte, 1024)
	_, err = conn.Read(b)
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("resp=%s", b))

	return nil
}

func (mpv *MpvSocket) Quit() error {
	log := slog.With("method", "MpvSocket.Quit")
	log.Info("stopping")

	conn, err := net.Dial("unix", mpv.sockFile)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(ipcCmds[quit] + "\n"))
	if err != nil {
		return err
	}

	b := make([]byte, 1024)
	_, err = conn.Read(b)
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("resp=%s", b))

	return nil
}

func (mpv *MpvSocket) getBaseCmd() string {
	res := baseCmd
	if runtime.GOOS == "windows" {
		res = baseCmdWindows
	}
	return res
}

func (mpv *MpvSocket) getSockFilePath() string {
	res := sockFile
	if runtime.GOOS == "windows" {
		res = sockFileWin
	}
	return fmt.Sprintf(res, os.Getpid())
}
