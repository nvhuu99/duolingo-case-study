package job_queue

import (
	"context"
	"errors"
)

var (
	ErrInvalidQueueName  = errors.New("invalid queue name")
	ErrMainQueueIsNotSet = errors.New("main queue is not set")
)

type TaskQueue interface {
	SetQueue(queue string)
	Declare() error
	Remove() error
}

type TaskProducer interface {
	SetQueue(queue string)
	Push(serializedTask string) error
}

type TaskConsumer interface {
	SetQueue(queue string)
	Consuming(ctx context.Context, handleFunc func(string) error) error
}
