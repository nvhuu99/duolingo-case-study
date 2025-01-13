package rabbitmq

import (
	"context"
	mqp "duolingo/lib/message-queue"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	defaultWriteTimeOut = 5 * time.Second
)
type RabbitMQPublisher struct {
	topic 		  *mqp.TopicInfo
	deliveryTag   uint64
	ctx   		  context.Context
	mu			  sync.Mutex
	timeOut		  time.Duration
}

func NewPublisher(ctx context.Context) *RabbitMQPublisher {
	publisher := RabbitMQPublisher{}
	publisher.ctx = ctx
	publisher.deliveryTag = 1
	publisher.timeOut = defaultWriteTimeOut
	return &publisher
}

// Sets the queue name to consume messages from.
func (publisher *RabbitMQPublisher) SetTopicInfo(topic *mqp.TopicInfo) {
	publisher.topic = topic
}

func (publisher *RabbitMQPublisher) Publish(message string) error {
	conn, ch, err := publisher.connect()
	if err != nil {
		return err
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	defer conn.Close()
	defer ch.Close()
	defer close(confirms)

	publisher.mu.Lock()
	topic := publisher.topic.Name
	routingKey, err := publisher.topic.Next()
	publisher.mu.Unlock()
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(publisher.ctx,
		topic,
		routingKey,
		true,  // mandatory (message must be routed to at least one queue)
		false, // immediate (queue message even when no consumers)
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		},
	)
	if err != nil {
		return mqp.NewMessageError(err.Error(), message, topic, routingKey)
	}

	select {
	case confirm := <-confirms:
		if confirm.Ack && confirm.DeliveryTag == publisher.deliveryTag {
			publisher.mu.Lock()
			publisher.deliveryTag++
			publisher.mu.Unlock()
			return nil
		}
		return mqp.NewMessageError("NACK (not acknowledged)", message, topic, routingKey)
	case <-time.After(publisher.timeOut):
		return mqp.NewMessageError("timeout waiting for publish confirmation", message, topic, routingKey)
	}
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
	if err := ch.Confirm(false); err != nil {
		return nil, nil, err
	}

	return conn, ch, nil
}
