package connection_manager

import "time"

type BaseConnectionArgs struct {
	connectionTimeout   time.Duration
	connectionRetryWait time.Duration

	operationReadTimeout  time.Duration
	operationWriteTimeout time.Duration
	operationRetryWait    time.Duration
}

func DefaultConnectionArgs() *BaseConnectionArgs {
	return &BaseConnectionArgs{
		connectionTimeout:     10 * time.Second,
		connectionRetryWait:   200 * time.Millisecond,
		operationReadTimeout:  15 * time.Second,
		operationWriteTimeout: 15 * time.Second,
		operationRetryWait:    200 * time.Millisecond,
	}
}

func (args *BaseConnectionArgs) GetConnectionTimeout() time.Duration {
	return args.connectionTimeout
}

func (args *BaseConnectionArgs) SetConnectionTimeout(value time.Duration) ConnectionArgs {
	args.connectionTimeout = value
	return args
}

func (args *BaseConnectionArgs) GetConnectionRetryWait() time.Duration {
	return args.connectionRetryWait
}

func (args *BaseConnectionArgs) SetConnectionRetryWait(value time.Duration) ConnectionArgs {
	args.connectionRetryWait = value
	return args
}

func (args *BaseConnectionArgs) GetOperationReadTimeout() time.Duration {
	return args.operationReadTimeout
}

func (args *BaseConnectionArgs) SetOperationReadTimeout(value time.Duration) ConnectionArgs {
	args.operationReadTimeout = value
	return args
}

func (args *BaseConnectionArgs) GetOperationWriteTimeout() time.Duration {
	return args.operationWriteTimeout
}

func (args *BaseConnectionArgs) SetOperationWriteTimeout(value time.Duration) ConnectionArgs {
	args.operationWriteTimeout = value
	return args
}

func (args *BaseConnectionArgs) GetOperationRetryWait() time.Duration {
	return args.operationRetryWait
}

func (args *BaseConnectionArgs) SetOperationRetryWait(value time.Duration) ConnectionArgs {
	args.operationRetryWait = value
	return args
}
