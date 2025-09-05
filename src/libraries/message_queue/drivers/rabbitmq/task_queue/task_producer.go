package task_queue

import (
	ctxt "context"
	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	events "duolingo/libraries/events/facade"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq"
	tq "duolingo/libraries/message_queue/task_queue"
	"fmt"
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

func (p *TaskProducer) Push(ctx ctxt.Context, serializedTask string) error {
	var err error

	evt := events.Start(ctx, fmt.Sprintf("task_queue.producer.push(%v)", p.queue), map[string]any{
		"task_queue": p.queue,
	})
	defer events.End(evt, true, err, nil)

	if p.queue == "" {
		err = tq.ErrInvalidQueueName
		return err
	}

	err = p.Publish(evt.Context(), p.queue, p.queue, serializedTask, nil)

	return err
}
