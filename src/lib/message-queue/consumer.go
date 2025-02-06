package messagequeue

type Consumer interface {
	Consume(handler func(string) ConsumerAction)
}