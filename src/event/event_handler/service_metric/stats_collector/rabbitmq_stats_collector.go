package metric

import (
	cnst "duolingo/constant"
	mq "duolingo/lib/message_queue/driver/rabbitmq"
	"duolingo/lib/metric"
	

	"github.com/google/uuid"
)

type RabbitMQStatsCollector struct {
	id    string
	published map[string]int
	publishLatency map[string]int64
	delivered map[string]int
	snapshots map[string][]*metric.Snapshot
}

func NewRabbitMQStatsCollector() *RabbitMQStatsCollector {
	c := new(RabbitMQStatsCollector)
	c.id = uuid.NewString()
	c.delivered = make(map[string]int)
	c.published = make(map[string]int)
	c.publishLatency = make(map[string]int64)
	c.snapshots = make(map[string][]*metric.Snapshot)
	return c
}

func (c *RabbitMQStatsCollector) SubscriberId() string {
	return c.id
}

func (c *RabbitMQStatsCollector) Notified(event string, data any) {
	switch event {
	case mq.EVT_CLIENT_ACTION_CONSUMED:
		if evt, ok := data.(*mq.ConsumeEvent); ok {
			c.delivered[evt.QueueName]++
		}
	case mq.EVT_CLIENT_ACTION_PUBLISHED:
		if evt, ok := data.(*mq.PublishEvent); ok {
			c.published[evt.QueueName]++
			c.publishLatency[evt.QueueName] = max(c.publishLatency[evt.QueueName], evt.Latency.Milliseconds())
		}
	}
}

func (c *RabbitMQStatsCollector) Capture() {
	defer func() {
		c.delivered = make(map[string]int)
		c.published = make(map[string]int)
		c.publishLatency = make(map[string]int64)
	}()
	for q, v := range c.delivered {
		c.snapshots["delivered"] = append(c.snapshots["delivered"], metric.NewSnapshot(float64(v),
			"queue", q, cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_ACCUMULATE))
	}
	for q, v := range c.published {
		c.snapshots["published"] = append(c.snapshots["published"], metric.NewSnapshot(float64(v), 
			"queue", q, cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_ACCUMULATE))
	}
	for q, v := range c.publishLatency {
		c.snapshots["publish_latency"] = append(c.snapshots["publish_latency"], metric.NewSnapshot(float64(v), 
			"queue", q, cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_MAXIMUM))
	}
}

func (c *RabbitMQStatsCollector) Collect() []*metric.DataPoint {
	defer func() {
		c.snapshots = make(map[string][]*metric.Snapshot)
	}()
	datapoints := []*metric.DataPoint{
		metric.RawDataPoint(c.snapshots["delivered"], "metric_target", cnst.METRIC_TARGET_RABBITMQ, "metric_name", cnst.METRIC_NAME_DELIVERED_RATE),
		metric.RawDataPoint(c.snapshots["published"], "metric_target", cnst.METRIC_TARGET_RABBITMQ, "metric_name", cnst.METRIC_NAME_PUBLISHED_RATE),
		metric.RawDataPoint(c.snapshots["publish_latency"], "metric_target", cnst.METRIC_TARGET_RABBITMQ, "metric_name", cnst.METRIC_NAME_PUBLISH_LATENCY),
	}

	return datapoints 
}
