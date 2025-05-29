package metric

import (
	cnst "duolingo/constant"
	mq "duolingo/lib/message_queue/driver/rabbitmq"
	"duolingo/lib/metric"
	"fmt"

	"github.com/google/uuid"
)

type RabbitMQStatsCollector struct {
	id    string
	published map[string]int
	delivered map[string]int
	snapshots map[string][]*metric.Snapshot
}

func NewRabbitMQStatsCollector() *RabbitMQStatsCollector {
	c := new(RabbitMQStatsCollector)
	c.id = uuid.NewString()
	c.delivered = make(map[string]int)
	c.published = make(map[string]int)
	c.snapshots = make(map[string][]*metric.Snapshot)
	return c
}

func (c *RabbitMQStatsCollector) SubscriberId() string {
	return c.id
}

func (c *RabbitMQStatsCollector) Notified(event string, data any) {
	evt, ok := data.(*mq.ClientActionEvent)
	if !ok {
		return
	}
	switch evt.Action {
	case mq.ConsumerAccept, mq.ConsumerReject:
		c.delivered[evt.QueueName]++
		fmt.Printf("rabbitmq stats collector message delivered: %v\n", c.delivered[evt.QueueName])
	case mq.PublisherPublished:
		c.published[evt.QueueName]++
	}
}

func (c *RabbitMQStatsCollector) Capture() {
	defer func() {
		c.delivered = make(map[string]int)
		c.published = make(map[string]int)
	}()
	for q, v := range c.delivered {
		c.snapshots["delivered"] = append(c.snapshots["delivered"], metric.NewSnapshot(float64(v),
			"queue", q, cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_ACCUMULATE))
	}
	for q, v := range c.published {
		c.snapshots["published"] = append(c.snapshots["published"], metric.NewSnapshot(float64(v), 
			"queue", q, cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_ACCUMULATE))
	}
}

func (c *RabbitMQStatsCollector) Collect() []*metric.DataPoint {
	defer func() {
		c.snapshots = make(map[string][]*metric.Snapshot)
	}()
	datapoints := []*metric.DataPoint{
		metric.RawDataPoint(c.snapshots["delivered"], "metric_target", cnst.METRIC_TARGET_RABBITMQ, "metric_name", cnst.METRIC_NAME_DELIVERED_RATE),
		metric.RawDataPoint(c.snapshots["published"], "metric_target", cnst.METRIC_TARGET_RABBITMQ, "metric_name", cnst.METRIC_NAME_PUBLISHED_RATE),
	}

	return datapoints 
}
