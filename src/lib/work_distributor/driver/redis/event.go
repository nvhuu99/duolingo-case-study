package redis

const (
	EVT_REDIS_COMMANDS_EXEC = "evt_redis_commands_exec"
	EVT_REDIS_LOCK_RELEASED = "evt_redis_lock_released"
)

type RedisCommandExecutedEvent struct {
	Count int8
}

type RedisLockReleasedEvent struct {
	WaitedTimeMs int64
	HeldTimeMs   int64
}
