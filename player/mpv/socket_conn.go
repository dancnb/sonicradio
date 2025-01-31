//go:build !windows

package mpv

import (
	"context"
	"net"
)

var (
	baseCmd  = "mpv"
	sockFile = "/tmp/mpvsocket.%d"
)

func getConn(ctx context.Context, addr string) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", addr)
	return conn, err
}
