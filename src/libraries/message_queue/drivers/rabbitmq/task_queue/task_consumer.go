package task_queue

import (
	"context"
	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	events "duolingo/libraries/events/facade"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq"
	tq "duolingo/libraries/message_queue/task_queue"
	"fmt"
)

type TaskConsumer struct {
	*driver.QueueConsumer

	queue string
}

func NewTaskConsumer(client *connection.RabbitMQClient) *TaskConsumer {
	return &TaskConsumer{
		QueueConsumer: &driver.QueueConsumer{
			Topology: driver.NewTopology(client),
		},
	}
}

func (c *TaskConsumer) SetQueue(queue string) {
	c.queue = queue
}

func (c *TaskConsumer) Consuming(
	ctx context.Context,
	handleFunc func(context.Context, string) error,
) error {
	if c.queue == "" {
		return tq.ErrInvalidQueueName
	}
	return c.QueueConsumer.Consuming(ctx, c.queue, func(
		receiveCtx context.Context,
		receiveMsg string,
	) (driver.ConsumeAction, error) {
		var err error

		func() {
			evt := events.Start(receiveCtx, fmt.Sprintf("task_queue.consumer.consume(%v)", c.queue), nil)
			defer events.End(evt, true, err, nil)

			err = handleFunc(evt.Context(), receiveMsg)
		}()

		return driver.ActionAccept, err
	})
}
