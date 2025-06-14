package connection_manager

import "time"

type BaseConnectionArgs struct {
	connectionTimeout time.Duration

	readTimeout  time.Duration
	writeTimeout time.Duration
	retryWait    time.Duration
}

func DefaultConnectionArgs() *BaseConnectionArgs {
	return &BaseConnectionArgs{
		connectionTimeout: 10 * time.Second,
		readTimeout:       15 * time.Second,
		writeTimeout:      15 * time.Second,
		retryWait:         200 * time.Millisecond,
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

func (args *BaseConnectionArgs) GetRetryWait() time.Duration {
	return args.retryWait
}

func (args *BaseConnectionArgs) SetRetryWait(value time.Duration) ConnectionArgs {
	args.retryWait = value
	return args
}
