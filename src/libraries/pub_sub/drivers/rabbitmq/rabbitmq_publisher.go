package rabbitmq

import (
	"context"
	"errors"

	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrPublisherMainTopicNotSet = errors.New("publisher main topic is not set")
)

type RabbitMQPublisher struct {
	*RabbitMQTopology
	mainTopic string
}

func NewRabbitMQPublisher(client *connection.RabbitMQClient) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		RabbitMQTopology: NewRabbitMQTopology(client),
	}
}

func (p *RabbitMQPublisher) SetMainTopic(topic string) {
	p.mainTopic = topic
}

func (p *RabbitMQPublisher) DeclareMainTopic() error {
	if p.mainTopic == "" {
		return ErrPublisherMainTopicNotSet
	}
	return p.DeclareTopic(p.mainTopic)
}

func (p *RabbitMQPublisher) RemoveMainTopic() error {
	if p.mainTopic == "" {
		return ErrPublisherMainTopicNotSet
	}
	return p.RemoveTopic(p.mainTopic)
}

func (p *RabbitMQPublisher) NotifyMainTopic(message string) error {
	if p.mainTopic == "" {
		return ErrPublisherMainTopicNotSet
	}
	return p.Notify(p.mainTopic, message)
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
			topic,
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
