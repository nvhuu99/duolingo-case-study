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
	keyToLocks []string,
	wait time.Duration,
	closure func(ctx context.Context, connection *redis_driver.Client) error,
) error {
	lock := NewDistributedLock(client, keyToLocks)
	if acquireErr := lock.AcquireLock(); acquireErr != nil {
		return acquireErr
	}
	defer lock.ReleaseLock()

	return client.ExecuteClosure(wait, closure)
}

func (client *RedisClient) ExecuteClosure(
	wait time.Duration,
	closure func(ctx context.Context, connection *redis_driver.Client) error,
) error {
	wrapper := func(ctx context.Context, conn any) error {
		converted, _ := conn.(*redis_driver.Client)
		return closure(ctx, converted)
	}
	return client.Client.ExecuteClosure(wait, wrapper)
}

func (client *RedisClient) GetConnection() *redis_driver.Client {
	connection := client.Client.GetConnection()
	if redisConn, ok := connection.(*redis_driver.Client); ok {
		return redisConn
	}
	return nil
}
