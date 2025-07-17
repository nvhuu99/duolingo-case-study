package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	*Topology
}

func (p *Publisher) Publish(
	topic string,
	key string,
	message string,
	headers map[string]string,
) error {
	timeout := p.GetWriteTimeout()
	return p.ExecuteClosure(timeout, func(ctx context.Context, ch *amqp.Channel) error {
		headerTable := amqp.Table{}
		if len(headers) != 0 {
			for k, v := range headers {
				headerTable[k] = v
			}
		}
		return ch.PublishWithContext(
			ctx,
			topic,
			key,
			true,  // message must be routed to at least one queue
			false, // queue message even when no consumers
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte(message),
				Headers:      headerTable,
			},
		)
	})
}
