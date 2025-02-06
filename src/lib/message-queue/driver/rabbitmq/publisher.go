package rabbitmq

import (
	"context"
	mq "duolingo/lib/message-queue"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	opts	*mq.PublisherOptions
	manager	mq.Manager

	id			string
	chanId		string
	confirm		chan amqp.Confirmation
	deliveryTag	uint64
	errChan		chan *mq.Error

	ctx		context.Context
	cancel	context.CancelFunc
}

func NewPublisher(ctx context.Context ,opts *mq.PublisherOptions) *RabbitMQPublisher {
	client := RabbitMQPublisher{}
	client.ctx, client.cancel = context.WithCancel(ctx)
	if opts == nil {
		opts = mq.DefaultPublisherOptions()
	}
	client.opts = opts
	
	return &client
}

func (client *RabbitMQPublisher) OnReConnected() {
}

func (client *RabbitMQPublisher) OnConnectionFailure(err *mq.Error) {
}

func (client *RabbitMQPublisher) OnClientFatalError(err *mq.Error) {
	client.terminate(err)
}

func (client *RabbitMQPublisher) NotifyError(ch chan *mq.Error) chan *mq.Error {
	client.errChan = ch
	return ch
}

func (client *RabbitMQPublisher) UseManager(manager mq.Manager) {
	client.id = manager.RegisterClient(client)
	client.manager = manager
}

func (client *RabbitMQPublisher) Publish(mssg string) *mq.Error {
	var publishErr *mq.Error
	topic := client.opts.Topic
	routingKey := client.opts.Dispatcher.Dispatch(mssg)
	writeDeadline := time.After(client.opts.WriteTimeOut)
	firstTry := true
	for {
		select {
		case <-client.ctx.Done():
			return publishErr
		case <-writeDeadline:
			return mq.NewError(mq.PublishTimeOutExceed, nil, topic, "", routingKey)
		default:
		}

		if ! firstTry {
			time.Sleep(client.opts.GraceTimeOut)
		}
		firstTry = false

		ch := client.getChannel()
		if ch == nil {
			publishErr = mq.NewError(mq.ConnectionFailure, nil, topic, "", routingKey)
			continue
		}
		
		err := ch.PublishWithContext(
			client.ctx,
			topic,
			routingKey,
			true,  // mandatory (message must be routed to at least one queue)
			true, // immediate (queue message even when no consumers)
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType: "text/plain",
				Body: []byte(mssg),
				Headers: amqp.Table{
					"message_id": uuid.New().String(),
				},
			},
		)
		if err != nil {
			publishErr = mq.NewError(mq.PublishFailure, err, topic, "", routingKey)
			continue
		}
		
		confirm := <-client.confirm

		if !confirm.Ack {
			publishErr = mq.NewError(mq.PublishNACK, nil, topic, "", routingKey)
			continue
		}

		if confirm.DeliveryTag != client.deliveryTag {
			publishErr = mq.NewError(mq.PublishConfirmFailure, nil, topic, "", routingKey)
			continue
		}

		client.deliveryTag++

		publishErr = nil
		
		break
	}

	return publishErr
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
		client.confirm = channel.NotifyPublish(make(chan amqp.Confirmation, 1))
		client.deliveryTag = 1
	}

	return channel
}

func (client *RabbitMQPublisher) sendErr(err *mq.Error) {
	if client.errChan != nil {
		client.errChan <- err
	}
}

func (client *RabbitMQPublisher) terminate(err *mq.Error) {
	go client.manager.UnRegisterClient(client.id)
	client.sendErr(err)
	client.cancel()
}