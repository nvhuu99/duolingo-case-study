package rabbitmq

import (
	"context"
	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	"duolingo/libraries/pub_sub"
	"errors"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQSubscriber struct {
	*RabbitMQTopology

	id string
}

func NewRabbitMQSubscriber(client *connection.RabbitMQClient) *RabbitMQSubscriber {
	return &RabbitMQSubscriber{
		RabbitMQTopology: NewRabbitMQTopology(client),
		id:             uuid.NewString(),
	}
}

func (sub *RabbitMQSubscriber) GetChannel() string {
	return sub.id
}

func (sub *RabbitMQSubscriber) Subscribe(topic string) error {
	return sub.DeclareQueue(sub.GetChannel(), "", topic)
}

func (sub *RabbitMQSubscriber) UnSubscribe(topic string) error {
	return sub.DeleteQueue(sub.GetChannel())
}

func (sub *RabbitMQSubscriber) Consuming(
	ctx context.Context,
	topic string,
	closure func(string) pub_sub.ConsumeAction,
) error {
	var deliveries <-chan amqp.Delivery
	var channel *amqp.Channel
	var fatalErr error

	deliveries, channel, fatalErr = sub.waitForDeliveriesChanReady(ctx)
	if fatalErr != nil {
		return fatalErr
	}
	defer func() {
		// For the next Consuming() to work, the channel must be closed and recreated.
		// All new messages will then flow to the most recently created deliveries channel
		// instead of the current one.
		channel.Close()
		sub.RenewConnection()
	}()

	var confirmationFailures = make(map[string]pub_sub.ConsumeAction)
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, connectionAlive := <-deliveries:
			if !connectionAlive {
				deliveries, channel, fatalErr = sub.waitForDeliveriesChanReady(ctx)
				if fatalErr != nil {
					return fatalErr
				}
				continue
			}
			// A duplicate message ID was received, indicating that the last
			// consume-action failed (as a result, the message has been automatically
			// requeued according to policy).
			// Since the message has already been processed, there's no need to
			// call the "closure" again; instead, simply retry the consume action.
			id, _ := delivery.Headers["message_id"].(string)
			if previousFailedAction, found := confirmationFailures[id]; found {
				retryErr := sub.handleConsumeAction(delivery, previousFailedAction)
				if retryErr == nil {
					delete(confirmationFailures, id)
				}
				continue
			}
			// Upon receiving a new message, first call the "closure" function,
			// then send the "confirmation" to the server (acknowledge, reject, etc.)
			// based on the "consume action" returned by the closure.
			consumeAction := closure(string(delivery.Body))
			actionErr := sub.handleConsumeAction(delivery, consumeAction)
			// If the confirmation fails, the failure will be recorded to be addressed later.
			if actionErr != nil {
				confirmationFailures[id] = consumeAction
			}
		}
	}
}

func (sub *RabbitMQSubscriber) waitForDeliveriesChanReady(
	ctx context.Context,
) (
	<-chan amqp.Delivery,
	*amqp.Channel,
	error,
) {
	for {
		select {
		case <-ctx.Done():
			return nil, nil, errors.New("fail to get deliveries channel due to context canceled")
		default:
		}
		if ch := sub.GetConnection(); ch != nil {
			deliveries, err := ch.Consume(
				sub.GetChannel(),
				"",    // consumer tag (empty string for auto-generated)
				false, // auto-ack (manual acknowledgment)
				false, // exclusive
				false, // no-local (allow messages from the same connection)
				false, // no-wait (wait for the queue to be created)
				nil,   // arguments (none)
			)
			if err == nil {
				return deliveries, ch, nil
			}
			if !sub.IsNetworkErr(err) {
				return nil, nil, err
			}
		}
		// Gracefully retry
		sub.NotifyNetworkFailure()
		time.Sleep(sub.GetRetryWait())
		continue
	}
}

func (sub *RabbitMQSubscriber) handleConsumeAction(
	delivery amqp.Delivery,
	action pub_sub.ConsumeAction,
) error {
	switch action {
	case pub_sub.ActionReject:
		return delivery.Reject(false)
	case pub_sub.ActionRequeue:
		return delivery.Reject(true)
	default:
		return delivery.Ack(false)
	}
}
