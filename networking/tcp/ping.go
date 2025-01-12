package tcp

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration
	select {
	case <-ctx.Done():
		return
	case interval = <-reset: // pulled initial interval off reset channel // 1
	default:
	}
	if interval <= 0 {
		interval = defaultPingInterval
	}

	timer := time.NewTimer(interval) // 2
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		select {
		case <-ctx.Done(): // 3
			return
		case newInterval := <-reset: // 4
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C: // 5
			if _, err := w.Write([]byte("ping")); err != nil {
				// If wanted, would use this case to keep track of any consecutive time-outs that occur while writing to the writer. -> could pass the context cancel function and call it in here if a threshold is reached
				// track and act on consecutive timeouts here
				return
			}
		}
		_ = timer.Reset(interval) // 6
	}
}
