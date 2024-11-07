package player

import (
	"context"
	"net"
)

var (
	baseCmd  = "mpv.exe"
	sockFile = `\\.\pipe\mpvsocket.%d`
)

func getConn(ctx context.Context, addr string) (net.Conn, error) {

	return nil, nil
}
