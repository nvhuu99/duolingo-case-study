package task_queue

import (
	"context"
	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq"
	tq "duolingo/libraries/message_queue/task_queue"
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
	handleFunc func(context.Context, string),
) error {
	if c.queue == "" {
		return tq.ErrInvalidQueueName
	}
	return c.QueueConsumer.Consuming(ctx, c.queue, func(
		ctx context.Context,
		msg string,
	) driver.ConsumeAction {
		handleFunc(ctx, msg)
		return driver.ActionAccept
	})
}
