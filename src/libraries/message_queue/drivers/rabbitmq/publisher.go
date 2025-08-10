package rabbitmq

import (
	"context"
	events "duolingo/libraries/events/facade"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	*Topology
}

func (p *Publisher) Publish(
	ctx context.Context,
	topic string,
	key string,
	message string,
	headers map[string]any,
) error {
	var err error

	if headers == nil {
		headers = make(map[string]any)
	}
	headerTable := amqp.Table(headers)

	evt := events.Start(
		ctx, fmt.Sprintf("mq.publisher.publish(%v)", topic),
		map[string]any{
			"routing_key":     key,
			"message_headers": headerTable,
		},
	)
	defer events.End(evt, true, err, nil)

	timeout := p.GetWriteTimeout()
	err = p.ExecuteClosure(evt.Context(), timeout, func(
		timeoutCtx context.Context,
		ch *amqp.Channel,
	) error {
		return ch.PublishWithContext(
			timeoutCtx,
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

	return err
}
