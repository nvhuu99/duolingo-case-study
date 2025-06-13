package redis

import (
	"context"

	"duolingo/libraries/connection_manager"
)

type RedisConnectionBuilder struct {
	*connection_manager.ConnectionBuilder
	ctx context.Context
}

func NewRedisConnectionBuilder(
	ctx context.Context,
	args *RedisConnectionArgs,
) *RedisConnectionBuilder {
	if args == nil {
		args = DefaultRedisConnectionArgs()
	}
	baseBuilder := connection_manager.NewConnectionBuilder(ctx)
	baseBuilder.SetConnectionArgs(args)
	baseBuilder.SetConnectionProxy(&RedisConnectionProxy{ctx: ctx})
	return &RedisConnectionBuilder{
		ctx:               ctx,
		ConnectionBuilder: baseBuilder,
	}
}

func (builder *RedisConnectionBuilder) BuildClientAndRegisterToManager() *RedisClient {
	redisArgs, ok := builder.GetConnectionArgs().(*RedisConnectionArgs)
	if !ok {
		panic(ErrConnectionArgsType)
	}
	client := builder.ConnectionBuilder.BuildClientAndRegisterToManager()
	redisClient := &RedisClient{
		Client:                  client,
		lockAcquireTimeout:      redisArgs.GetLockAcquireTimeout(),
		lockAcquireRetryWaitMin: redisArgs.GetLockAcquireRetryWaitMin(),
		lockAcquireRetryWaitMax: redisArgs.GetLockAcquireRetryWaitMax(),
		lockTTL:                 redisArgs.GetLockTTL(),
	}

	return redisClient
}
