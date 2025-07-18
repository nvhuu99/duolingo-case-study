package task_queue

import (
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

func (q *TaskQueue) Declare() error {
	if q.queue == "" {
		return tq.ErrInvalidQueueName
	}
	var declareErr error
	declareErr = q.DeclareExchange(
		driver.
			DefaultExchangeOpts(q.queue).
			IsType(driver.DirectExchange).
			IsPersistent(),
	)
	if declareErr == nil {
		declareErr = q.DeclareQueue(
			driver.DefaultQueueOpts(q.queue).IsPersistent(),
			driver.NewQueueBinding(q.queue).Add(q.queue, q.queue),
		)
	}
	return declareErr
}

func (q *TaskQueue) Remove() error {
	if q.queue == "" {
		return nil
	}
	var err error
	if err = q.DeleteExchange(q.queue); err == nil {
		err = q.DeleteQueue(q.queue)
	}
	return err
}
