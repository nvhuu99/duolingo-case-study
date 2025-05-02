package metric

import (
	mq "duolingo/lib/message_queue/driver/rabbitmq"
	"github.com/google/uuid"
)

type RabbitMQStats struct {
	Delivered uint `json:"delivered"`
	Published uint `json:"published"`
}

type RabbitMQStatsCollector struct {
	id string
	stats map[string]*RabbitMQStats
}

func NewRabbitMQStatsCollector() *RabbitMQStatsCollector {
	c := new(RabbitMQStatsCollector)
	c.id = uuid.NewString()
	c.stats = make(map[string]*RabbitMQStats)
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
	if _, exists := c.stats[evt.QueueName]; !exists {
		c.stats[evt.QueueName] = new(RabbitMQStats)
	}
	switch evt.Action {
	case mq.ConsumerAccept, mq.ConsumerReject:
		c.stats[evt.QueueName].Delivered++
	case mq.PublisherPublished:
		c.stats[evt.QueueName].Published++
	}
}

func (c *RabbitMQStatsCollector) Capture() any {
	stats := make(map[string]RabbitMQStats)
	for q, s := range c.stats {
		stats[q] = *s
	}
	return stats
}
