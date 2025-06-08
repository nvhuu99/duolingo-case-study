package connection_manager

import "time"

type ConnectArgs struct {
	URI               string
	ConnectionTimeout time.Duration
	ConnectionRetryWait time.Duration
	OperationRetryWait time.Duration
	OperationReadTimeout time.Duration
	OperationWriteTimeout time.Duration
}