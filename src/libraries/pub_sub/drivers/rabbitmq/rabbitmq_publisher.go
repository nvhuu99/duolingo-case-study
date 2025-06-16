package rabbitmq

import (
	"context"

	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	*RabbitMQTopology
}

func NewRabbitMQPublisher(client *connection.RabbitMQClient) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		RabbitMQTopology: NewRabbitMQTopology(client),
	}
}

func (p *RabbitMQPublisher) DeclareTopic(topic string) error {
	return p.DeclareExchange(topic)
}

func (p *RabbitMQPublisher) RemoveTopic(topic string) error {
	return p.DeleteExchange(topic)
}

func (p *RabbitMQPublisher) Notify(topic string, message string) error {
	timeout := p.GetWriteTimeout()
	return p.ExecuteClosure(timeout, func(ctx context.Context, ch *amqp.Channel) error {
		return ch.PublishWithContext(
			ctx,
			topic,
			"",    // fanout publish
			true,  // mandatory (message must be routed to at least one queue)
			false, // immediate (queue message even when no consumers)
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte(message),
				Headers: amqp.Table{
					"message_id": uuid.NewString(),
				},
			},
		)
	})
}
