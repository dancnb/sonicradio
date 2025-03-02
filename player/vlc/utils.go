//go:build !windows

package vlc

const baseCmd = "vlc"

func GetBaseCmd() string {
	return baseCmd
}

// func getConn(ctx context.Context) (net.Conn, error) {
// 	sockFile, err := getUnixSocket(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	start := time.Now()
// loop:
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil, ErrCtxCancel
// 		case <-time.After(socketTimeout):
// 			return nil, ErrSocketFileTimeout
// 		default:
// 			if _, err := os.Stat(sockFile); os.IsNotExist(err) {
// 				time.Sleep(socketSleepRetry)
// 			} else {
// 				break loop
// 			}
// 		}
// 	}
// 	slog.Info(fmt.Sprintf("vlc unix socket file created after %v", time.Since(start)))

// 	var d net.Dialer
// 	conn, err := d.DialContext(ctx, "unix", sockFile)
// 	return conn, err
// }

// func getUnixSocket(ctx context.Context) (string, error) {
// 	log := slog.With("method", "vlcCmd")
// 	sockFile := fmt.Sprintf(connAddr, os.Getpid())
// 	args := slices.Clone(baseArgs)
// 	args = append(args, sockFile)
// 	cmd := exec.CommandContext(ctx, GetBaseCmd(), args...)
// 	if errors.Is(cmd.Err, exec.ErrDot) {
// 		cmd.Err = nil
// 	} else if cmd.Err != nil {
// 		log.Error("vlc cmd error", "error", cmd.Err.Error())
// 		return "", cmd.Err
// 	}
// 	// cmd.Stderr = &bytes.Buffer{}
// 	// cmd.Stdin = &bytes.Buffer{}
// 	err := cmd.Start()
// 	if err != nil {
// 		log.Error("vlc cmd start", "error", err)
// 		return "", err
// 	}
// 	log.Info("vlc cmd started", "pid", cmd.Process.Pid)
// 	return sockFile, nil
// }
