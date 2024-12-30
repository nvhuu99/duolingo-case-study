package rabbitmq

import (
	"context"
	mqp "duolingo/lib/message-queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	topic mqp.TopicInfo
	ctx   context.Context
}

func NewPublisher(ctx context.Context) mqp.MessagePublisher {
	publisher := RabbitMQPublisher{}
	publisher.ctx = ctx

	return &publisher
}

// Sets the queue name to consume messages from.
func (publisher *RabbitMQPublisher) SetTopicInfo(topic mqp.TopicInfo) {
	publisher.topic = topic
}

func (publisher *RabbitMQPublisher) Publish(message string) error {
	conn, ch, err := publisher.connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	topic := publisher.topic.Name
	routingKey, err := publisher.topic.Next()
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(publisher.ctx,
		topic,
		routingKey,
		true,  // mandatory (message must be routed to at least one queue)
		false, // immediate (message will be delivered immediately)
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		},
	)

	if err != nil {
		return mqp.NewMessageError(err.Error(), message, topic, routingKey)
	}

	return nil
}

func (publisher *RabbitMQPublisher) connect() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(publisher.topic.ConnectionString)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}
