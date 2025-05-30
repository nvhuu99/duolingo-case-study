package metric

import (
	cnst "duolingo/constant"
	mongo "duolingo/repository/campaign_user/event"
	"duolingo/lib/metric"
	

	"github.com/google/uuid"
)

type MongoStatsCollector struct {
	id    string
	commands int
	latency int64
	snapshots map[string][]*metric.Snapshot
}

func NewMongoStatsCollector() *MongoStatsCollector {
	c := new(MongoStatsCollector)
	c.id = uuid.NewString()
	c.snapshots = make(map[string][]*metric.Snapshot)
	return c
}

func (c *MongoStatsCollector) SubscriberId() string {
	return c.id
}

func (c *MongoStatsCollector) Notified(event string, data any) {
	switch event {
	case mongo.EVT_MONGODB_QUERY:
		if evt, ok := data.(*mongo.MongoDBQueryEvent); ok {
			c.commands++
			c.latency = max(c.latency, evt.Latency.Milliseconds())
		}
	}
}

func (c *MongoStatsCollector) Capture() {
	defer func() {
		c.commands = 0
		c.latency = 0
	}()
	c.snapshots["cmd"] = append(c.snapshots["cmd"], metric.NewSnapshot(float64(c.commands),
		cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_ACCUMULATE))
	c.snapshots["latency"] = append(c.snapshots["latency"], metric.NewSnapshot(float64(c.latency),
		cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_MAXIMUM))
}

func (c *MongoStatsCollector) Collect() []*metric.DataPoint {
	defer func() {
		c.snapshots = make(map[string][]*metric.Snapshot)
	}()
	datapoints := []*metric.DataPoint{
		metric.RawDataPoint(c.snapshots["cmd"], "metric_target", cnst.METRIC_TARGET_MONGO, "metric_name", cnst.METRIC_NAME_QUERY_RATE),
		metric.RawDataPoint(c.snapshots["latency"], "metric_target", cnst.METRIC_TARGET_MONGO, "metric_name", cnst.METRIC_NAME_QUERY_LATENCY),
	}

	return datapoints 
}
