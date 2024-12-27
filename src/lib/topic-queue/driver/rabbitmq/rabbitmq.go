package rabbitmq

import (
	"context"
	"duolingo/lib/helper-functions"
	tq "duolingo/lib/topic-queue"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type mqStatus string

const (
	statusOpened mqStatus = "opened"
	statusClosed mqStatus = "closed"

	errQueueClosed    = "either the instance is closed or not yet opened"
	errExchangeNotSet = "exchange name must be specified before publishing messages"
	errQueueNotSet    = "queue name must be specified before consuming messages"
)

type RabbitMQ struct {
	uri      string
	exchange string
	pattern  string
	queue    string
	status   mqStatus
	err      error

	conn *amqp.Connection
	ch   *amqp.Channel

	ctx     context.Context
	mu      sync.Mutex
	timeOut time.Duration
}

func NewQueue(ctx context.Context) tq.MessageQueue {
	mq := RabbitMQ{}
	mq.ctx = ctx
	mq.status = statusClosed

	return &mq
}

// Sets the URI for the RabbitMQ connection.
func (mq *RabbitMQ) UseConnectionString(uri string) {
	if mq.status != statusOpened {
		mq.uri = uri
	}
}

// Sets the connection parameters (host, port, user, password) and generates the URI.
func (mq *RabbitMQ) UseConnection(host string, port string, user string, pwd string) {
	if mq.status != statusOpened {
		pwd = url.QueryEscape(pwd)
		mq.uri = fmt.Sprintf("amqp://%v:%v@%v:%v/", user, pwd, host, port)
	}
}

// Sets the timeout duration for consuming messages.
func (mq *RabbitMQ) SetReadTimeOut(timeOut time.Duration) {
	if mq.status != statusOpened {
		mq.timeOut = timeOut
	}
}

// Sets the exchange and routing pattern for publishing messages.
func (mq *RabbitMQ) SetPublishRoute(exchange string, pattern string) {
	if mq.status != statusOpened {
		mq.exchange = exchange
		mq.pattern = pattern
	}
}

// Sets the queue name to consume messages from.
func (mq *RabbitMQ) SetConsumeQueue(queue string) {
	if mq.status != statusOpened {
		mq.queue = queue
	}
}

// Establishes a connection to RabbitMQ and opens a channel for communication.
func (mq *RabbitMQ) Open() error {
	if mq.status == statusOpened {
		return nil
	}
	conn, err := amqp.Dial(mq.uri)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	mq.conn = conn
	mq.ch = ch
	mq.status = statusOpened

	return nil
}

// Close closes the connection and channel to RabbitMQ, and sets the status to closed.
func (mq *RabbitMQ) Close() {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if mq.status == statusOpened {
		mq.status = statusClosed
		mq.conn.Close()
		mq.ch.Close()
	}
}

// Sends a message to the RabbitMQ exchange with the specified routing pattern.
func (mq *RabbitMQ) Publish(message string) error {
	mq.mu.Lock()
	if mq.status != statusOpened {
		mq.mu.Unlock()
		return errors.New(errQueueClosed)
	}
	if mq.exchange == "" {
		mq.mu.Unlock()
		return errors.New(errExchangeNotSet)
	}

	// Publishing the message
	err := mq.ch.PublishWithContext(mq.ctx,
		mq.exchange,
		mq.pattern,
		true,  // mandatory (message must be routed to at least one queue)
		false, // immediate (message will be delivered immediately)
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		},
	)

	mq.mu.Unlock()

	// Handle the returned message
	go func() {
		for r := range mq.ch.NotifyReturn(make(chan amqp.Return)) {
			mq.mu.Lock()
			mq.err = tq.NewMessageError(r.ReplyText, string(r.Body), r.Exchange, r.RoutingKey)
			mq.mu.Unlock()
		}
	}()

	return err
}

// Listens to the specified queue and processes messages using the provided handler
// The handler should return false as a signal to stop listening.
func (mq *RabbitMQ) Consume(handler func(string) bool) error {
	mq.mu.Lock()
	if mq.queue == "" {
		mq.mu.Unlock()
		return errors.New(errQueueNotSet)
	}

	// Consuming messages from the queue
	msgs, err := mq.ch.Consume(
		mq.queue,
		"",       // consumer tag (empty string for auto-generated)
		false,    // auto-ack (manual acknowledgment)
		false,    // exclusive
		false,    // no-local (allow messages from the same connection)
		false,    // no-wait (wait for the queue to be created)
		nil,      // arguments (none)
	)
	mq.mu.Unlock()

	if err != nil {
		return err
	}

	// Handling the messages
	go func() {
		for d := range msgs {
			// If the handler takes too long, ACK is not sent, and the message is requeued
			done := helper.OperationDeadline(mq.ctx, mq.timeOut, nil, func() { d.Ack(false) })
			check := handler(string(d.Body))
			done <- true
			// If the handler returns false, stop consuming further messages
			if !check {
				break
			}
		}
	}()

	return nil
}

func (mq *RabbitMQ) Error() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	err := mq.err

	return err 
}