package redismanager

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func acquireLock(ctx context.Context, rdb *redis.Client, ttl time.Duration, val string, keys ...string) error {
	script := redis.NewScript(`
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

	const maxRetries = 1000
	for i := 0; i < maxRetries; i++ {
		result, err := script.Run(ctx, rdb, keys, val, ttl).Result()
		if err == nil && result == int64(1) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return errors.New("batch manager: reached maximum number of retries to acquire the lock")
}

func releaseLock(ctx context.Context, rdb *redis.Client, val string, keys ...string) {
	script := redis.NewScript(`
		local keys = KEYS
        local lock_value = ARGV[1]

        for i, key in ipairs(keys) do
            if redis.call("GET", key) == lock_value then
                redis.call("DEL", key)
            end
        end
	`)
	script.Run(ctx, rdb, keys, val)
}

