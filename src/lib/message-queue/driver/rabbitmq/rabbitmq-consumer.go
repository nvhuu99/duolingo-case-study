package rabbitmq

import (
	"context"
	mqp "duolingo/lib/message-queue"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultReadTimeout = 30 * time.Second
	errQueueNotSet     = "queue name must be specified before consuming messages"
)

type RabbitMQConsumer struct {
	queue *mqp.QueueInfo

	ctx     context.Context
	timeOut time.Duration
}

func NewConsumer(ctx context.Context) *RabbitMQConsumer {
	mq := RabbitMQConsumer{}
	mq.ctx = ctx
	mq.timeOut = defaultReadTimeout

	return &mq
}

// Sets the timeout duration for consuming messages.
func (mq *RabbitMQConsumer) SetReadTimeOut(timeOut time.Duration) {
	mq.timeOut = timeOut
}

// Sets the queue name to consume messages from.
func (mq *RabbitMQConsumer) SetQueueInfo(queue *mqp.QueueInfo) {
	mq.queue = queue
}

// Listens to the specified queue and processes messages using the provided handler
// The handler should return false as a signal to stop listening.
func (mq *RabbitMQConsumer) Consume(handler func(string) bool) error {
	if mq.queue.QueueName == "" {
		return errors.New(errQueueNotSet)
	}

	conn, ch, err := mq.connect()
	if err != nil {
		return err
	}

	// Consuming messages from the queue
	msgs, err := ch.Consume(
		mq.queue.QueueName,
		"",       // consumer tag (empty string for auto-generated)
		false,    // auto-ack (manual acknowledgment)
		false,    // exclusive
		false,    // no-local (allow messages from the same connection)
		false,    // no-wait (wait for the queue to be created)
		nil,      // arguments (none)
	)

	if err != nil {
		return err
	}

	// Handling the messages
	go func() {
		defer conn.Close()
		defer ch.Close()

		consuming := func(d amqp.Delivery) {
			if !handler(string(d.Body)) {
				d.Reject(true)
			} else {
				d.Ack(false)
			}
		}

		for {
			select {
			case <-mq.ctx.Done():
				return
			case d := <- msgs:
				consuming(d)
			}
		}
	}()

	return nil
}

func (topic *RabbitMQConsumer) connect() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(topic.queue.ConnectionString)
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