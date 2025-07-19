package connection_manager

import "time"

type BaseConnectionArgs struct {
	connectionTimeout time.Duration

	readTimeout  time.Duration
	writeTimeout time.Duration
	retryWait    time.Duration
}

func DefaultConnectionArgs() *BaseConnectionArgs {
	connectionTimeout := 30 * time.Second
	return &BaseConnectionArgs{
		connectionTimeout: connectionTimeout,
		readTimeout:       connectionTimeout + 15*time.Second,
		writeTimeout:      connectionTimeout + 15*time.Second,
		retryWait:         100 * time.Millisecond,
	}
}

func (args *BaseConnectionArgs) GetConnectionTimeout() time.Duration {
	return args.connectionTimeout
}

func (args *BaseConnectionArgs) SetConnectionTimeout(value time.Duration) ConnectionArgs {
	args.connectionTimeout = value
	return args
}

func (args *BaseConnectionArgs) GetReadTimeout() time.Duration {
	return args.readTimeout
}

func (args *BaseConnectionArgs) SetReadTimeout(value time.Duration) ConnectionArgs {
	args.readTimeout = value
	return args
}

func (args *BaseConnectionArgs) GetWriteTimeout() time.Duration {
	return args.writeTimeout
}

func (args *BaseConnectionArgs) SetWriteTimeout(value time.Duration) ConnectionArgs {
	args.writeTimeout = value
	return args
}

func (args *BaseConnectionArgs) GetOperationTimeout() time.Duration {
	return max(args.GetReadTimeout(), args.GetWriteTimeout())
}

func (args *BaseConnectionArgs) GetRetryWait() time.Duration {
	return args.retryWait
}

func (args *BaseConnectionArgs) SetRetryWait(value time.Duration) ConnectionArgs {
	args.retryWait = value
	return args
}
