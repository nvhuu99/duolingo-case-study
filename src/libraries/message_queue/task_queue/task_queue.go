package task_queue

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
	Declare(ctx context.Context) error
	Remove(ctx context.Context) error
}

type TaskProducer interface {
	SetQueue(queue string)
	Push(ctx context.Context, serializedTask string) error
}

type TaskConsumer interface {
	SetQueue(queue string)
	Consuming(ctx context.Context, handleFunc func(context.Context, string)) error
}
