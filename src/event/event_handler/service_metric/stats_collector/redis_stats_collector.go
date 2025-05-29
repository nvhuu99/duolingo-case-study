package metric

import (
	cnst "duolingo/constant"
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
	c.snapshots["command_rate"] = append(c.snapshots["command_rate"], metric.NewSnapshot(float64(c.commandCount), 
		cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_ACCUMULATE))
	c.snapshots["lock_waited"] = append(c.snapshots["lock_waited"], metric.NewSnapshot(float64(c.lockWaited), 
		cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_MAXIMUM))
	c.snapshots["lock_held"] = append(c.snapshots["lock_held"], metric.NewSnapshot(float64(c.lockHeld), 
		cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_MAXIMUM))
}

func (c *RedisStatsCollector) Collect() []*metric.DataPoint {
	defer func() { 
		c.snapshots = make(map[string][]*metric.Snapshot) 
	}()
	datapoints := []*metric.DataPoint{
		metric.RawDataPoint(c.snapshots["command_rate"], "metric_target", cnst.METRIC_TARGET_REDIS, "metric_name", cnst.METRIC_NAME_REDIS_CMD_RATE),
		metric.RawDataPoint(c.snapshots["lock_waited"], "metric_target", cnst.METRIC_TARGET_REDIS, "metric_name", cnst.METRIC_NAME_REDIS_LOCK_WAITED),
		metric.RawDataPoint(c.snapshots["lock_held"], "metric_target", cnst.METRIC_TARGET_REDIS, "metric_name", cnst.METRIC_NAME_REDIS_LOCK_HELD),
	}
	return datapoints
}