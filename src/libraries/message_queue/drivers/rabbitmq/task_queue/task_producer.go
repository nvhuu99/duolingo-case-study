package task_queue

import (
	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq"
	tq "duolingo/libraries/message_queue/task_queue"
)

type TaskProducer struct {
	*driver.Publisher

	queue string
}

func NewTaskProducer(client *connection.RabbitMQClient) *TaskProducer {
	return &TaskProducer{
		Publisher: &driver.Publisher{
			Topology: driver.NewTopology(client),
		},
	}
}

func (p *TaskProducer) SetQueue(queue string) {
	p.queue = queue
}

func (p *TaskProducer) Push(serializedTask string) error {
	if p.queue == "" {
		return tq.ErrInvalidQueueName
	}
	return p.Publish(p.queue, p.queue, serializedTask, nil)
}
