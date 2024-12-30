package messagequeue

type MessageConsumer interface {
	SetQueueInfo(queue QueueInfo)
	Consume(hanlder func(string) bool) error
}