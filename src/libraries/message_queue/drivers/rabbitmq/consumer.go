package rabbitmq

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueConsumer struct {
	*Topology
}

func (c *QueueConsumer) Consuming(
	ctx context.Context,
	queue string,
	closure func(context.Context, string) ConsumeAction,
) error {
	var deliveries <-chan amqp.Delivery
	var channel *amqp.Channel
	var fatalErr error

	deliveries, channel, fatalErr = c.waitForDeliveriesChanReady(ctx, queue)
	if fatalErr != nil {
		return fatalErr
	}
	defer func() {
		// For the next Consuming() to work, the channel must be closed and recreated.
		// All new messages will then flow to the most recently created deliveries channel
		// instead of the current one.
		channel.Close()
		c.RenewConnection()
	}()

	var confirmationFailures = make(map[string]ConsumeAction)
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, connectionAlive := <-deliveries:
			if !connectionAlive {
				deliveries, channel, fatalErr = c.waitForDeliveriesChanReady(ctx, queue)
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
				retryErr := c.handleConsumeAction(delivery, previousFailedAction)
				if retryErr == nil {
					delete(confirmationFailures, id)
				}
				continue
			}
			// Upon receiving a new message, first call the "closure" function,
			// then send the "confirmation" to the server (acknowledge, reject, etc.)
			// based on the "consume action" returned by the closure.
			consumeAction := closure(ctx, string(delivery.Body))
			actionErr := c.handleConsumeAction(delivery, consumeAction)
			// If the confirmation fails, the failure will be recorded to be addressed later.
			if actionErr != nil {
				confirmationFailures[id] = consumeAction
			}
		}
	}
}

func (c *QueueConsumer) waitForDeliveriesChanReady(
	ctx context.Context,
	queue string,
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
		if ch := c.GetConnection(); ch != nil {
			deliveries, err := ch.Consume(
				queue,
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
			if !c.IsNetworkErr(err) {
				return nil, nil, err
			}
		}
		// Gracefully retry
		c.NotifyNetworkFailure()
		time.Sleep(c.GetRetryWait())
		continue
	}
}

func (c *QueueConsumer) handleConsumeAction(
	delivery amqp.Delivery,
	action ConsumeAction,
) error {
	switch action {
	case ActionReject:
		return delivery.Reject(false)
	case ActionRequeue:
		return delivery.Reject(true)
	default:
		return delivery.Ack(false)
	}
}
