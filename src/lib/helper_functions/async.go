package helper

import (
	"context"
	"time"
)

// Starts a timer, returns a done channel that can be used to mark that the operation is finished.
// If done is triggered before the deadline, calls onDone(), otherwise calls onFail()
func OperationDeadline(ctx context.Context, duration time.Duration, onFail func(), onDone func()) chan bool {
	ctx, cancel := context.WithTimeout(ctx, duration*1000)
	done := make(chan bool, 1)

	go func() {
		defer func() {
			close(done)
			cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				if onFail != nil {
					onFail()
				}
				return
			case <-done:
				if onDone != nil {
					onDone()
				}
				return
			}
		}
	}()

	return done
}
