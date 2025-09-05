package task_queue

import (
	"context"
	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq"
	tq "duolingo/libraries/message_queue/task_queue"
)

type TaskQueue struct {
	*driver.Topology
	queue string
}

func NewTaskQueue(client *connection.RabbitMQClient) *TaskQueue {
	return &TaskQueue{
		Topology: driver.NewTopology(client),
	}
}

func (q *TaskQueue) SetQueue(queue string) {
	q.queue = queue
}

func (q *TaskQueue) Declare(ctx context.Context) error {
	if q.queue == "" {
		return tq.ErrInvalidQueueName
	}
	var declareErr error
	declareErr = q.DeclareExchange(
		ctx,
		driver.
			DefaultExchangeOpts(q.queue).
			IsType(driver.DirectExchange).
			IsPersistent(),
	)
	if declareErr == nil {
		declareErr = q.DeclareQueue(
			ctx,
			driver.DefaultQueueOpts(q.queue).IsPersistent(),
			driver.NewQueueBinding(q.queue).Add(q.queue, q.queue),
		)
	}
	return declareErr
}

func (q *TaskQueue) Remove(ctx context.Context) error {
	if q.queue == "" {
		return nil
	}
	var err error
	if err = q.DeleteExchange(ctx, q.queue); err == nil {
		err = q.DeleteQueue(ctx, q.queue)
	}
	return err
}
