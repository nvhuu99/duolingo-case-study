package redis

import (
	"context"
	"duolingo/libraries/connection_manager"
	"time"

	redis_driver "github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*connection_manager.Client

	lockAcquireTimeout      time.Duration
	lockAcquireRetryWaitMin time.Duration
	lockAcquireRetryWaitMax time.Duration
	lockTTL                 time.Duration
}

func (client *RedisClient) ExecuteClosureWithLocks(
	ctx context.Context,
	keyToLocks []string,
	timeout time.Duration,
	closure func(timeoutCtx context.Context, connection *redis_driver.Client) error,
) error {
	lock := NewDistributedLock(client, keyToLocks)
	if acquireErr := lock.AcquireLock(ctx); acquireErr != nil {
		return acquireErr
	}
	defer lock.ReleaseLock(ctx)

	return client.ExecuteClosure(ctx, timeout, closure)
}

func (client *RedisClient) ExecuteClosure(
	ctx context.Context,
	timeout time.Duration,
	closure func(timeoutCtx context.Context, connection *redis_driver.Client) error,
) error {
	wrapper := func(ctx context.Context, conn any) error {
		converted, _ := conn.(*redis_driver.Client)
		return closure(ctx, converted)
	}
	return client.Client.ExecuteClosure(ctx, timeout, wrapper)
}

func (client *RedisClient) GetConnection() *redis_driver.Client {
	connection := client.Client.GetConnection()
	if redisConn, ok := connection.(*redis_driver.Client); ok {
		return redisConn
	}
	return nil
}
