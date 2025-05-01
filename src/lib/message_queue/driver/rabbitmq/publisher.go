package rabbitmq

import (
	"context"
	mq "duolingo/lib/message_queue"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	opts    *mq.PublisherOptions
	manager mq.Manager

	id          string
	name        string
	chanId      string
	confirm     chan amqp.Confirmation
	deliveryTag uint64

	ctx    context.Context
	cancel context.CancelFunc
}

func NewPublisher(name string, ctx context.Context) *RabbitMQPublisher {
	client := RabbitMQPublisher{}
	client.ctx, client.cancel = context.WithCancel(ctx)
	client.opts = mq.DefaultPublisherOptions()
	client.name = name

	return &client
}

func (client *RabbitMQPublisher) WithOptions(opts *mq.PublisherOptions) *mq.PublisherOptions {
	if opts == nil {
		client.opts = mq.DefaultPublisherOptions()
	} else {
		client.opts = opts
	}
	return client.opts
}

func (client *RabbitMQPublisher) ResetConnection() {
}

func (client *RabbitMQPublisher) UseManager(manager mq.Manager) {
	client.id = manager.RegisterClient(client.name, client)
	client.manager = manager
}

func (client *RabbitMQPublisher) Publish(mssg string) error {
	var publishErr error
	topic := client.opts.Topic
	routingKey := client.opts.Dispatcher.Dispatch(mssg)
	writeDeadline := time.After(client.opts.WriteTimeOut)
	firstTry := true

	defer func() {
		if publishErr == nil {
			if manager, ok := client.manager.(*RabbitMQManager); ok {
				manager.opts.EventPublisher.Notify(EVT_ON_CLIENT_ACTION, &ClientActionEvent{
					ClientName: client.name,
					QueueName: client.opts.Topic,
					Action: PublisherPublished,
				})
			}
		}
	}()

	for {
		select {
		case <-client.ctx.Done():
			return publishErr
		case <-writeDeadline:
			return errors.New(mq.ErrMessages[mq.ERR_PUBLISH_TIMEOUT_EXCEED])
		default:
		}

		if !firstTry {
			time.Sleep(client.opts.GraceTimeOut)
		}
		firstTry = false

		ch := client.getChannel()
		if ch == nil {
			publishErr = errors.New(mq.ErrMessages[mq.ERR_CONNECTION_FAILURE])
			continue
		}

		err := ch.PublishWithContext(
			client.ctx,
			topic,
			routingKey,
			true,  // mandatory (message must be routed to at least one queue)
			false, // immediate (queue message even when no consumers)
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte(mssg),
				Headers: amqp.Table{
					"message_id": uuid.New().String(),
				},
			},
		)

		if err != nil {
			publishErr = fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_PUBLISH_FAILURE], err)
			continue
		}

		confirm := <-client.confirm

		if !confirm.Ack {
			publishErr = errors.New(mq.ErrMessages[mq.ERR_PUBLISH_NACK])
			continue
		}

		if confirm.DeliveryTag != client.deliveryTag {
			publishErr = errors.New(mq.ErrMessages[mq.ERR_PUBLISH_CONFIRM_FAILURE])
			continue
		}

		client.deliveryTag++

		publishErr = nil

		break
	}

	return nil
}

func (client *RabbitMQPublisher) getChannel() *amqp.Channel {
	ch, chId := client.manager.GetClientConnection(client.id)
	channel, ok := ch.(*amqp.Channel)
	if !ok || channel == nil {
		return nil
	}
	if client.chanId != chId {
		if err := channel.Confirm(false); err != nil {
			return nil
		}
		client.chanId = chId
		client.confirm = channel.NotifyPublish(make(chan amqp.Confirmation, 10))
		client.deliveryTag = 1
	}

	return channel
}
