package vlc

import (
	"context"
	"fmt"
	"net"
	"slices"
)

var (
	baseArgs = []string{"-I rc", "--rc-quiet", "--no-video", "--volume-step 12.8", "--gain 1.0"}
	connArg  = "--rc-host"
	connAddr = "localhost:%d"
)

const baseCmd = "vlc"

func GetBaseCmd() string {
	return baseCmd
}

func getArgs() []string {
	res := slices.Clone(baseArgs)
	res = append(res, fmt.Sprintf(connArg))
	return res
}

// func getFreePort() (port int, err error) {
// 	a, _ := net.ResolveTCPAddr("tcp", "localhost:0")
// 	var l *net.TCPListener
// 	if l, err = net.ListenTCP("tcp", a); err != nil {
// 		defer l.Close()
// 		return l.Addr().(*net.TCPAddr).Port, nil
// 	}
// 	return
// }

func getConn(ctx context.Context, addr string) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	return conn, err
}
