package rabbitmq

import (
	"context"
	mq "duolingo/lib/message_queue"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	opts       *mq.ConsumerOptions
	manager    mq.Manager
	deliveries <-chan amqp.Delivery
	reset      chan bool

	id   string
	name string

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

func NewConsumer(name string, ctx context.Context) *RabbitMQConsumer {
	client := RabbitMQConsumer{}
	client.ctx, client.cancel = context.WithCancel(ctx)
	client.opts = mq.DefaultConsumerOptions()
	client.reset = make(chan bool, 1)
	client.name = name

	return &client
}

func (client *RabbitMQConsumer) WithOptions(opts *mq.ConsumerOptions) *mq.ConsumerOptions {
	if opts == nil {
		client.opts = mq.DefaultConsumerOptions()
	} else {
		client.opts = opts
	}
	return client.opts
}

func (client *RabbitMQConsumer) ResetConnection() {
	client.reset <- true
}

func (client *RabbitMQConsumer) UseManager(manager mq.Manager) {
	client.id = manager.RegisterClient(client.name, client)
	client.manager = manager
}

func (client *RabbitMQConsumer) Consume(done <-chan bool, handler func([]byte) mq.ConsumerAction) {
	confirmationFailures := make(map[string]mq.ConsumerAction, 1)

	client.mu.RLock()
	deliveries := client.deliveries
	client.mu.RUnlock()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case <-client.ctx.Done():
				return
			case <-client.reset:
				client.resetDeliveries()
				client.mu.RLock()
				deliveries = client.deliveries
				client.mu.RUnlock()
			case d, ok := <-deliveries:
				// Received an empty message, most of the time
				// the cause would be the connection lost.
				// This message will be skipped.
				if len(d.Body) == 0 {
					if ok {
						client.action(d, mq.ConsumerAccept)
					} else {
						time.Sleep(client.opts.GraceTimeOut)
					}
					continue
				}
				// Received an duplication message id, the cause is very much likely
				// that the last confirmation has failed, and the message is requeued.
				// Since the "consumer" has already processed the message, needs not to
				// call the consumer "handler" again.
				// Instead, only send the "confirmation".
				id, _ := d.Headers["message_id"].(string)
				if failure, found := confirmationFailures[id]; found {
					if ok, _ := client.action(d, failure); ok {
						delete(confirmationFailures, id)
					}
					time.Sleep(client.opts.GraceTimeOut)
					continue
				}
				// Received a new message, first call the consumer "handler",
				// then send the "confirmation" to the server (ack, reject, etc.).
				act := handler(d.Body)
				result, _ := client.action(d, act)
				// If the confirmation failed, the message will be requeued automatically
				if !result {
					// The confirmation failure is recorded to be handled again later.
					confirmationFailures[id] = act
					time.Sleep(client.opts.GraceTimeOut)
					continue
				}
			}
		}
	}()
	wg.Wait()
}

func (client *RabbitMQConsumer) getChannel() *amqp.Channel {
	ch, _ := client.manager.GetClientConnection(client.id)
	channel, ok := ch.(*amqp.Channel)
	if ch == nil || !ok {
		return nil
	}

	return channel
}

func (client *RabbitMQConsumer) resetDeliveries() {
	var deliveries <-chan amqp.Delivery
	var err error
	firstTry := true
	for {
		select {
		case <-client.ctx.Done():
			return
		default:
		}

		if !firstTry {
			time.Sleep(client.opts.GraceTimeOut)
		}
		firstTry = false

		ch := client.getChannel()
		if ch == nil {
			continue
		}

		deliveries, err = ch.Consume(
			client.opts.Queue,
			"",    // consumer tag (empty string for auto-generated)
			false, // auto-ack (manual acknowledgment)
			false, // exclusive
			false, // no-local (allow messages from the same connection)
			false, // no-wait (wait for the queue to be created)
			nil,   // arguments (none)
		)

		if err == nil {
			client.mu.Lock()
			client.deliveries = deliveries
			client.mu.Unlock()
			return
		}
	}
}

func (client *RabbitMQConsumer) action(d amqp.Delivery, act mq.ConsumerAction) (bool, error) {
	var err error
	var clientAct ClientAction
	switch act {
	case mq.ConsumerRequeue:
		err = d.Reject(true)
		clientAct = ConsumerRequeue
	case mq.ConsumerReject:
		err = d.Reject(false)
		clientAct = ConsumerReject
	default:
		err = d.Ack(false)
		clientAct = ConsumerAccept
	}
	if err == nil {
		if manager, ok := client.manager.(*RabbitMQManager); ok {
			manager.opts.EventPublisher.Notify(EVT_ON_CLIENT_ACTION, &ClientActionEvent{
				ClientName: client.name,
				QueueName:  client.opts.Queue,
				Action:     clientAct,
			})
		}
	}
	return err == nil, err
}
