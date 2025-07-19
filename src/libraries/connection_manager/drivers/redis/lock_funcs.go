package redis

import (
	"context"
	"errors"
	"time"

	redis_driver "github.com/redis/go-redis/v9"
)

func getLockKeysForResourceKeys(resourceKeys []string) []string {
	lockKeys := make([]string, len(resourceKeys))
	for i := range resourceKeys {
		lockKeys[i] = "distributed_lock:" + resourceKeys[i]
	}
	return lockKeys
}

func acquireLock(
	ctx context.Context,
	rdb *redis_driver.Client,
	lockVal string,
	resourceKeys []string,
	ttl time.Duration,
) error {
	lockKeys := getLockKeysForResourceKeys(resourceKeys)
	script := redis_driver.NewScript(`
		local keys = KEYS
		local lock_value = ARGV[1]
		local ttl = tonumber(ARGV[2])

		for i, key in ipairs(keys) do
			if redis.call("EXISTS", key) == 1 then
				return 0 -- Lock failed
			end
		end

		for i, key in ipairs(keys) do
			redis.call("SET", key, lock_value, "PX", ttl)
		end

		return 1 -- Lock succeeded
	`)
	result, lockErr := script.Run(ctx, rdb, lockKeys, lockVal, ttl.Milliseconds()).Result()
	if result == int64(0) {
		if lockErr == nil {
			lockErr = errors.New("the resources have already been locking")
		}
		return lockErr
	}
	return nil
}

func releaseLock(
	ctx context.Context,
	rdb *redis_driver.Client,
	lockVal string,
	resourceKeys []string,
) error {
	lockKeys := getLockKeysForResourceKeys(resourceKeys)
	script := redis_driver.NewScript(`
		local keys = KEYS
        local lock_value = ARGV[1]

        for i, key in ipairs(keys) do
            if redis.call("GET", key) == lock_value then
                redis.call("DEL", key)
            end
        end
	`)
	result, lockErr := script.Run(ctx, rdb, lockKeys, lockVal).Result()
	if result == int64(0) {
		if lockErr == nil {
			lockErr = errors.New("unknown error")
		}
		return lockErr
	}
	return nil
}
