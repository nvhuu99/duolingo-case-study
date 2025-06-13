package connection_manager

import "time"

type ConnectionArgs interface {
	GetConnectionTimeout() time.Duration
	GetConnectionRetryWait() time.Duration
	GetOperationReadTimeout() time.Duration
	GetOperationWriteTimeout() time.Duration
	GetOperationRetryWait() time.Duration

	SetConnectionTimeout(time.Duration) ConnectionArgs
	SetConnectionRetryWait(time.Duration) ConnectionArgs
	SetOperationReadTimeout(time.Duration) ConnectionArgs
	SetOperationWriteTimeout(time.Duration) ConnectionArgs
	SetOperationRetryWait(time.Duration) ConnectionArgs
}
