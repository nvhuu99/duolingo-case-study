package rabbitmq

import (
	"context"
	mq "duolingo/lib/message-queue"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	opts		*mq.ConsumerOptions
	manager		mq.Manager
	deliveries	<-chan amqp.Delivery
	errChan		chan *mq.Error
	id			string

	ctx		context.Context
	cancel	context.CancelFunc
}

func NewConsumer(ctx context.Context ,opts *mq.ConsumerOptions) *RabbitMQConsumer {
	client := RabbitMQConsumer{}
	client.ctx, client.cancel = context.WithCancel(ctx)
	if opts == nil {
		opts = mq.DefaultConsumerOptions()
	}
	client.opts = opts
	
	return &client
}

func (client *RabbitMQConsumer) OnConnectionFailure(err *mq.Error) {
}

func (client *RabbitMQConsumer) OnClientFatalError(err *mq.Error) {
	client.terminate(err)
}

func (client *RabbitMQConsumer) NotifyError(ch chan *mq.Error) chan *mq.Error {
	client.errChan = ch
	return ch
}

func (client *RabbitMQConsumer) OnReConnected() {
	client.openDeliveriesChannel()
}

func (client *RabbitMQConsumer) UseManager(manager mq.Manager) {
	client.id = manager.RegisterClient(client)
	client.manager = manager
	client.openDeliveriesChannel()
}

func (client *RabbitMQConsumer) Consume(handler func(string) mq.ConsumerAction) {
	confirmationFailures := make(map[string]mq.ConsumerAction)
	for {
		select {
		case <-client.ctx.Done():
			return
		case d := <-client.deliveries:
			// Received an empty message, most of the time 
			// the cause would be the connection lost. 
			// This message will be skipped.
			if len(d.Body) == 0 {
				// Gracefully wait for the connection to be ready again, 
				// It does not matter the ACK is received by the server.
				time.Sleep(client.opts.GraceTimeOut)
				continue
			}
			// Received an duplication message id, the cause is very much likely
			// that the last confirmation has failed, and the message is requeued.
			// Since the "consumer" has already processed the message, need not to
			// call the consumer "handler" again. 
			// Instead, only send the "confirmation".
			id, _ := d.Headers["message_id"].(string)
			if failure, found := confirmationFailures[id]; found {
				if ok := client.action(d, failure); !ok {
					delete(confirmationFailures, id)
				}
				continue
			}
			// Received a new message, first call the consumer "handler",
			// then send the "confirmation" to the server (ack, reject, etc.).
			act := handler(string(d.Body))
			ok := client.action(d, act)
			// If the confirmation failed, the message will be requeued automatically
			if ! ok {
				// The confirmation failure is recorded to be handled again later.
				confirmationFailures[id] = act
				// The error might occur because a connection issue,
				// therefore, gracefully wait for the connection to be ready.
				time.Sleep(client.opts.GraceTimeOut) 
			}
		}
	}
}

func (client *RabbitMQConsumer) getChannel() *amqp.Channel {
	ch, _ := client.manager.GetClientConnection(client.id)
	channel, ok := ch.(*amqp.Channel)
	if !ok || channel == nil {
		return nil
	}

	return channel
}

func (client *RabbitMQConsumer) openDeliveriesChannel() {
	var deliveries <-chan amqp.Delivery
	var err error
	firstTry := true
	for {
		select {
		case <-client.ctx.Done():
			return
		default:
		}

		if ! firstTry {
			time.Sleep(client.opts.GraceTimeOut)
		}
		firstTry = false

		ch := client.getChannel()
		if ch == nil {
			continue
		}

		deliveries, err = ch.Consume(
			client.opts.Queue,
			"",       // consumer tag (empty string for auto-generated)
			false,    // auto-ack (manual acknowledgment)
			false,    // exclusive
			false,    // no-local (allow messages from the same connection)
			false,    // no-wait (wait for the queue to be created)
			nil,      // arguments (none)
		)

		if err == nil {
			client.deliveries = deliveries
			return
		}
	}
}

func (client *RabbitMQConsumer) action(d amqp.Delivery, act mq.ConsumerAction) bool {
	var err error
	switch act {
	case mq.ConsumerAccept:
		err = d.Ack(false)
	case mq.ConsumerRejectAndRequeue:
		err = d.Reject(true)
	case mq.ConsumerRejectAndDrop:
		err = d.Reject(false)
	default:
		return true
	}
	return err == nil
}

func (client *RabbitMQConsumer) sendErr(err *mq.Error) {
	if client.errChan != nil {
		client.errChan <- err
	}
}

func (client *RabbitMQConsumer) terminate(err *mq.Error) {
	go client.manager.UnRegisterClient(client.id)
	client.sendErr(err)
	client.cancel()
}