package redis

import (
	"context"
	"net/url"
	"time"

	"duolingo/libraries/connection_manager"
)

type RedisConnectionBuilder struct {
	connection_manager.ConnectionBuilder

	driver *RedisConnectionProxy

	Host    string
	Port    string
	User    string
	Passwod string

	LockAcquireTimeout      time.Duration
	LockAcquireRetryWaitMin time.Duration
	LockAcquireRetryWaitMax time.Duration
	LockTTL                 time.Duration

	ctx context.Context
}

func NewRedisConnectionBuilder(ctx context.Context) *RedisConnectionBuilder {
	builder := &RedisConnectionBuilder{
		ctx:                     ctx,
		LockAcquireTimeout:      10 * time.Second,
		LockAcquireRetryWaitMin: 10 * time.Millisecond,
		LockAcquireRetryWaitMax: 50 * time.Millisecond,
		LockTTL:                 2 * time.Second,
	}
	builder.ConnectionBuilder = *connection_manager.NewConnectionBuilder(ctx)
	builder.driver = &RedisConnectionProxy{ctx: ctx}
	return builder
}

func (builder *RedisConnectionBuilder) SetHost(host string) *RedisConnectionBuilder {
	builder.Host = host
	return builder
}

func (builder *RedisConnectionBuilder) SetPort(port string) *RedisConnectionBuilder {
	builder.Port = port
	return builder
}

func (builder *RedisConnectionBuilder) SetCredentials(user string, password string) *RedisConnectionBuilder {
	builder.User = url.QueryEscape(user)
	builder.Passwod = url.QueryEscape(password)
	return builder
}

func (builder *RedisConnectionBuilder) SetLockAcquireTimeout(duration time.Duration) *RedisConnectionBuilder {
	builder.LockAcquireTimeout = duration
	return builder
}

func (builder *RedisConnectionBuilder) SetLockAcquireRetryWait(min time.Duration, max time.Duration) *RedisConnectionBuilder {
	builder.LockAcquireRetryWaitMin = min
	builder.LockAcquireRetryWaitMax = max
	return builder
}

func (builder *RedisConnectionBuilder) SetLockTTL(duration time.Duration) *RedisConnectionBuilder {
	builder.LockTTL = duration
	return builder
}

func (builder *RedisConnectionBuilder) BuildConnectionManager() (*connection_manager.ConnectionManager, error) {
	args := &RedisConnectionArgs{
		ConnectionArgs: connection_manager.ConnectionArgs{
			ConnectionTimeout:     builder.ConnectionTimeout,
			ConnectionRetryWait:   builder.ConnectionRetryWait,
			OperationRetryWait:    builder.OperationRetryWait,
			OperationReadTimeout:  builder.OperationReadTimeout,
			OperationWriteTimeout: builder.OperationWriteTimeout,
		},
		Host:     builder.Host,
		Port:     builder.Port,
		User:     builder.User,
		Password: builder.Passwod,
	}
	builder.driver.SetConnectionArgsWithPanicOnValidationErr(args)
	builder.SetConnectionDriver(builder.driver)

	return builder.ConnectionBuilder.BuildConnectionManager()
}

func (builder *RedisConnectionBuilder) BuildClientAndRegisterToManager() (*RedisClient, error) {
	client, err := builder.ConnectionBuilder.BuildClientAndRegisterToManager()
	if err != nil {
		return nil, err
	}

	redisClient := &RedisClient{
		Client:                  *client,
		lockAcquireTimeout:      builder.LockAcquireTimeout,
		lockAcquireRetryWaitMin: builder.LockAcquireRetryWaitMin,
		lockAcquireRetryWaitMax: builder.LockAcquireRetryWaitMax,
		lockTTL:                 builder.LockTTL,
	}

	return redisClient, nil
}
