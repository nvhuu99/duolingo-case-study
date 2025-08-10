package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"time"

	events "duolingo/libraries/events/facade"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueConsumer struct {
	*Topology
}

func (c *QueueConsumer) Consuming(
	ctx context.Context,
	queue string,
	processFunc func(context.Context, string) (ConsumeAction, error),
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
			if prevFailedAction, found := confirmationFailures[id]; found {
				retryErr := c.handleConsumeAction(
					context.Background(), 
					queue, 
					delivery, 
					prevFailedAction,
				)
				if retryErr == nil {
					delete(confirmationFailures, id)
				}
				continue
			}
			// Upon receiving a new message, first call the "closure" function,
			// then send the "confirmation" to the server (acknowledge, reject, etc.)
			// based on the "consume action" returned by the closure.
			// CONTEXT PROPAGATION
			func() {
				var action ConsumeAction
				var ackErr error
				var processErr error

				evt := events.Start(
					ctx,
					fmt.Sprintf("mq.consumer.receive(%v)", queue),
					map[string]any{
						"message_headers": delivery.Headers,
					},
				)
				defer func() {
					events.End(evt, true, c.firstError(processErr, ackErr) , nil)
				}()

				action, processErr = processFunc(evt.Context(), string(delivery.Body))
				ackErr = c.handleConsumeAction(evt.Context(), queue, delivery, action)
				// If the confirmation fails, the failure will be recorded 
				// to be addressed later.
				if ackErr != nil {
					confirmationFailures[id] = action
				}
			}()
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
	ctx context.Context,
	queue string,
	delivery amqp.Delivery,
	action ConsumeAction,
) error {
	var err error
	
	evt := events.Start(ctx, fmt.Sprintf("mq.consumer.ack(%v, %v)", queue, action), nil)
	defer events.End(evt, true, err, nil)

	switch action {
	case ActionReject:
		err = delivery.Reject(false)
	case ActionRequeue:
		err = delivery.Reject(true)
	default:
		err = delivery.Ack(false)
	}

	return err
}

func (c *QueueConsumer) firstError(err ...error) error {
	if len(err) == 0 {
		return nil
	}
	return err[0]
}