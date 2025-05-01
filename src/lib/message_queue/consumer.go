package message_queue

type Consumer interface {
	Consume(done <-chan bool, handler func([]byte) ConsumerAction)
}
