package writer

import "time"

type LogWriter interface {
	WithBuffering(sizeMb int, maxCount int) LogWriter
	WithRotation(interval time.Duration) LogWriter
	WithFlushInterval(interval time.Duration) LogWriter
	Write(log *Writable)
}
