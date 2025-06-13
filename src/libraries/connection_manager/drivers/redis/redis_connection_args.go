package redis

import (
	"duolingo/libraries/connection_manager"
	"time"
)

type RedisConnectionArgs struct {
	*connection_manager.BaseConnectionArgs

	uri                     string
	host                    string
	port                    string
	user                    string
	password                string
	lockAcquireTimeout      time.Duration
	lockAcquireRetryWaitMin time.Duration
	lockAcquireRetryWaitMax time.Duration
	lockTTL                 time.Duration
}

func DefaultRedisConnectionArgs() *RedisConnectionArgs {
	baseArgs := connection_manager.DefaultConnectionArgs()
	redisArgs := &RedisConnectionArgs{
		BaseConnectionArgs:      baseArgs,
		host:                    "127.0.0.1",
		port:                    "6379",
		user:                    "",
		password:                "",
		lockAcquireTimeout:      20 * time.Second,
		lockAcquireRetryWaitMin: 10 * time.Millisecond,
		lockAcquireRetryWaitMax: 100 * time.Millisecond,
		lockTTL:                 15 * time.Second,
	}
	return redisArgs
}

func (r *RedisConnectionArgs) GetURI() string {
	return r.uri
}

func (r *RedisConnectionArgs) SetURI(uri string) *RedisConnectionArgs {
	r.uri = uri
	return r
}

func (r *RedisConnectionArgs) GetHost() string {
	return r.host
}

func (r *RedisConnectionArgs) SetHost(host string) *RedisConnectionArgs {
	r.host = host
	return r
}

func (r *RedisConnectionArgs) GetPort() string {
	return r.port
}

func (r *RedisConnectionArgs) SetPort(port string) *RedisConnectionArgs {
	r.port = port
	return r
}

func (r *RedisConnectionArgs) GetUser() string {
	return r.user
}

func (r *RedisConnectionArgs) SetUser(user string) *RedisConnectionArgs {
	r.user = user
	return r
}

func (r *RedisConnectionArgs) GetPassword() string {
	return r.password
}

func (r *RedisConnectionArgs) SetPassword(password string) *RedisConnectionArgs {
	r.password = password
	return r
}

func (r *RedisConnectionArgs) GetLockAcquireTimeout() time.Duration {
	return r.lockAcquireTimeout
}

func (r *RedisConnectionArgs) SetLockAcquireTimeout(timeout time.Duration) *RedisConnectionArgs {
	r.lockAcquireTimeout = timeout
	return r
}

func (r *RedisConnectionArgs) GetLockAcquireRetryWaitMin() time.Duration {
	return r.lockAcquireRetryWaitMin
}

func (r *RedisConnectionArgs) SetLockAcquireRetryWaitMin(wait time.Duration) *RedisConnectionArgs {
	r.lockAcquireRetryWaitMin = wait
	return r
}

func (r *RedisConnectionArgs) GetLockAcquireRetryWaitMax() time.Duration {
	return r.lockAcquireRetryWaitMax
}

func (r *RedisConnectionArgs) SetLockAcquireRetryWaitMax(wait time.Duration) *RedisConnectionArgs {
	r.lockAcquireRetryWaitMax = wait
	return r
}

func (r *RedisConnectionArgs) GetLockTTL() time.Duration {
	return r.lockTTL
}

func (r *RedisConnectionArgs) SetLockTTL(ttl time.Duration) *RedisConnectionArgs {
	r.lockTTL = ttl
	return r
}
