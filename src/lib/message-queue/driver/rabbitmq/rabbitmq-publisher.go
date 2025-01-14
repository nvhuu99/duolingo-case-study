package rabbitmq

import (
	"context"
	mqp "duolingo/lib/message-queue"
	"errors"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	defaultWriteTimeOut = 500 * time.Second

	errPublisherNotOpened = "publisher must be opened before publishing"
)

type RabbitMQPublisher struct {
	topic 		  *mqp.TopicInfo
	deliveryTag   uint64
	ctx   		  context.Context
	mu			  sync.Mutex
	timeOut		  time.Duration

	conn    *amqp.Connection
	ch      *amqp.Channel
	confirm chan amqp.Confirmation
}

func NewPublisher(ctx context.Context) *RabbitMQPublisher {
	publisher := RabbitMQPublisher{}
	publisher.ctx = ctx
	publisher.timeOut = defaultWriteTimeOut
	return &publisher
}

func (publisher *RabbitMQPublisher) Connect() error {
	conn, err := amqp.Dial(publisher.topic.ConnectionString)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return err
	}
	
	publisher.conn = conn
	publisher.ch = ch
	publisher.confirm = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	publisher.deliveryTag = 1

	return nil
}

func (publisher *RabbitMQPublisher) Disconnect() {
	if publisher.ch != nil {
		publisher.ch.Close()
	}
	if publisher.conn != nil {
		publisher.conn.Close()
	}
}

// Sets the queue name to consume messages from.
func (publisher *RabbitMQPublisher) SetTopicInfo(topic *mqp.TopicInfo) {
	publisher.topic = topic
}

func (publisher *RabbitMQPublisher) Publish(message string) error {
	if publisher.conn == nil || publisher.ch == nil {
		return errors.New(errPublisherNotOpened)
	}

	publisher.mu.Lock()
	topic := publisher.topic.Name
	routingKey, err := publisher.topic.Next()
	publisher.mu.Unlock()
	if err != nil {
		return err
	}

	err = publisher.ch.PublishWithContext(publisher.ctx,
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
	case confirm := <-publisher.confirm:
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

