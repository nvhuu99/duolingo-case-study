package metric

import (
	mq "duolingo/lib/message_queue/driver/rabbitmq"
	"duolingo/lib/metric"

	"github.com/google/uuid"
)

type RabbitMQStatsCollector struct {
	id    string
	published map[string]int
	delivered map[string]int
	depth map[string]int
	snapshots map[string][]*metric.Snapshot
}

func NewRabbitMQStatsCollector() *RabbitMQStatsCollector {
	c := new(RabbitMQStatsCollector)
	c.id = uuid.NewString()
	c.delivered = make(map[string]int)
	c.published = make(map[string]int)
	c.snapshots = make(map[string][]*metric.Snapshot)
	c.depth = make(map[string]int)
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
		if c.depth[evt.QueueName] > 0 {
			c.depth[evt.QueueName]--
		}
	case mq.PublisherPublished:
		c.published[evt.QueueName]++
		c.depth[evt.QueueName]++
	}
}

func (c *RabbitMQStatsCollector) Capture() {
	defer func() {
		c.delivered = make(map[string]int)
		c.published = make(map[string]int)
		c.depth = make(map[string]int)
	}()
	for q, v := range c.delivered {
		c.snapshots["delivered"] = append(c.snapshots["delivered"], metric.NewSnapshot(float64(v), "queue", q))
	}
	for q, v := range c.published {
		c.snapshots["published"] = append(c.snapshots["published"], metric.NewSnapshot(float64(v), "queue", q))
	}
	for q, v := range c.depth {
		c.snapshots["depth"] = append(c.snapshots["depth"], metric.NewSnapshot(float64(v), "queue", q))
	}
}

func (c *RabbitMQStatsCollector) Collect() []*metric.DataPoint {
	defer func() {
		c.snapshots = make(map[string][]*metric.Snapshot)
	}()
	datapoints := []*metric.DataPoint{
		metric.RawDataPoint(c.snapshots["delivered"], "metric_target", "rabbitmq", "metric_name", "delivered"),
		metric.RawDataPoint(c.snapshots["published"], "metric_target", "rabbitmq", "metric_name", "published"),
		metric.RawDataPoint(c.snapshots["depth"], "metric_target", "rabbitmq", "metric_name", "queue_depth"),
	}

	return datapoints 
}
