package metric

import (
	redis "duolingo/lib/work_distributor/driver/redis"
	"github.com/google/uuid"
)

type RedisCommandStats struct {
	Count int8 `json:"count"`
}

type RedisLockStats struct {
	WaitedMs int64 `json:"waited_ms"`
	HeldMs   int64 `json:"held_ms"`
}

type RedisStats struct {
	LockStats    []*RedisLockStats  `json:"lock_stats"`
	CommandStats *RedisCommandStats `json:"command_stats"`
}

type RedisStatsCollector struct {
	id    string
	stats *RedisStats
}

func NewRedisStatsCollector() *RedisStatsCollector {
	c := new(RedisStatsCollector)
	c.id = uuid.NewString()
	c.stats = &RedisStats{
		LockStats:    []*RedisLockStats{},
		CommandStats: new(RedisCommandStats),
	}
	return c
}

func (c *RedisStatsCollector) SubscriberId() string {
	return c.id
}

func (c *RedisStatsCollector) Notified(event string, data any) {
	switch event {
	case redis.EVT_REDIS_COMMANDS_EXEC:
		if evt, ok := data.(*redis.RedisCommandExecutedEvent); ok {
			c.stats.CommandStats.Count += evt.Count
		}
	case redis.EVT_REDIS_LOCK_RELEASED:
		if evt, ok := data.(*redis.RedisLockReleasedEvent); ok {
			c.stats.LockStats = append(c.stats.LockStats, &RedisLockStats{
				WaitedMs: evt.WaitedTimeMs,
				HeldMs:   evt.HeldTimeMs,
			})
		}
	}
}

func (c *RedisStatsCollector) Capture() any {
	stats := *c.stats
	c.stats.CommandStats.Count = 0
	c.stats.LockStats = []*RedisLockStats{}
	return stats
}
