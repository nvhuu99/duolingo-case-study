package connection_manager

import "time"

type ConnectionArgs interface {
	GetConnectionTimeout() time.Duration
	SetConnectionTimeout(time.Duration) ConnectionArgs

	GetReadTimeout() time.Duration
	SetReadTimeout(time.Duration) ConnectionArgs

	GetWriteTimeout() time.Duration
	SetWriteTimeout(time.Duration) ConnectionArgs

	GetRetryWait() time.Duration
	SetRetryWait(time.Duration) ConnectionArgs
}
