package metric

import (
	"duolingo/lib/metric"
	redis "duolingo/lib/work_distributor/driver/redis"

	"github.com/google/uuid"
)

type RedisStatsCollector struct {
	id    string
	commandCount int64
	lockWaited int64
	lockHeld int64
	snapshots map[string][]*metric.Snapshot
}

func NewRedisStatsCollector() *RedisStatsCollector {
	c := new(RedisStatsCollector)
	c.id = uuid.NewString()
	c.snapshots = make(map[string][]*metric.Snapshot)
	return c
}

/* Implement Event lib/event/subscriber interface */

func (c *RedisStatsCollector) SubscriberId() string {
	return c.id
}

func (c *RedisStatsCollector) Notified(event string, data any) {
	switch event {
	case redis.EVT_REDIS_COMMANDS_EXEC:
		if evt, ok := data.(*redis.RedisCommandExecutedEvent); ok {
			c.commandCount += int64(evt.Count)
		}
	case redis.EVT_REDIS_LOCK_RELEASED:
		if evt, ok := data.(*redis.RedisLockReleasedEvent); ok {
			c.lockWaited = max(c.lockWaited, evt.WaitedTimeMs)
			c.lockHeld = max(c.lockHeld, evt.HeldTimeMs)
		}
	}
}

/* Implement lib/metric/collector interface */

func (c *RedisStatsCollector) Capture() {
	defer func() {
		c.commandCount = 0
		c.lockWaited = 0
		c.lockHeld = 0
	}()
	c.snapshots["command_count"] = append(c.snapshots["command_count"], metric.NewSnapshot(float64(c.commandCount)))
	c.snapshots["lock_waited"] = append(c.snapshots["lock_waited"], metric.NewSnapshot(float64(c.lockWaited)))
	c.snapshots["lock_held"] = append(c.snapshots["lock_held"], metric.NewSnapshot(float64(c.lockHeld)))
}

func (c *RedisStatsCollector) Collect() []*metric.DataPoint {
	defer func() { 
		c.snapshots = make(map[string][]*metric.Snapshot) 
	}()
	datapoints := []*metric.DataPoint{
		metric.RawDataPoint(c.snapshots["command_count"], "service", "redis", "target", "command_count"),
		metric.RawDataPoint(c.snapshots["lock_waited"], "service", "redis", "target", "lock_waited"),
		metric.RawDataPoint(c.snapshots["lock_held"], "service", "redis", "target", "lock_held"),
	}
	return datapoints
}