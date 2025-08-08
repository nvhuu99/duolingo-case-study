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
	headers map[string]string,
) error {
	var err error 

	_, evt := events.Start(
		ctx, fmt.Sprintf("mq.publisher.publish(%v)", topic),
		map[string]any{
			"routing_key": key,
			"publisher_name": "message_input_publisher",
		},
	)
	defer func() {
		events.End(evt, err == nil, err, nil)
	}()

	timeout := p.GetWriteTimeout()
	err = p.ExecuteClosure(timeout, func(ctx context.Context, ch *amqp.Channel) error {
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

	return err
}
