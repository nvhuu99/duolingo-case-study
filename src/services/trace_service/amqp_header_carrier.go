package trace_service

import "github.com/rabbitmq/amqp091-go"

type AMQPHeadersCarrier amqp091.Table

func (c AMQPHeadersCarrier) Get(key string) string {
	if val, ok := c[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func (c AMQPHeadersCarrier) Set(key, value string) {
	c[key] = value
}

func (c AMQPHeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}