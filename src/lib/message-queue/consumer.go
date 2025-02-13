package messagequeue

type Consumer interface {
	Consume(done <-chan bool, handler func(string) ConsumerAction)
}