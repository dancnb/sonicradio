package mpv

import (
	"context"
	"net"
	"time"

	"gopkg.in/natefinch/npipe.v2"
)

var (
	baseCmd     = "mpv.exe"
	sockFile    = `\\.\pipe\mpvsocket.%d`
	dialTimeout = 2 * time.Second
)

func getConn(ctx context.Context, addr string) (net.Conn, error) {
	conn, err := npipe.DialTimeout(addr, dialTimeout)
	return conn, err
}
