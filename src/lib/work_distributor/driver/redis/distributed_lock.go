package redis

import (
	"context"
	"errors"
	// "math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

func acquireLock(ctx context.Context, rdb *redis.Client, ttl time.Duration, val string, keys ...string) error {
	var lockErr error
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
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return lockErr
		case <-timeout:
			return errors.New("failed to acquire the lock before timeout")
		default:
			var result any
			result, lockErr = script.Run(ctx, rdb, keys, val, ttl.Milliseconds()).Result()
			if result == int64(1) {
				return nil
			}
			// minWait := 10
			// maxWait := 100
			// wait := rand.Intn(maxWait-minWait+1) + minWait
			wait := 2
			time.Sleep(time.Duration(wait) * time.Millisecond)
		}
	}
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
