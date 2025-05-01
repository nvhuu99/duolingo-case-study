package message_queue

type Publisher interface {
	Publish(message string) error
}
