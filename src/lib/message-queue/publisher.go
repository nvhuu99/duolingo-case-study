package messagequeue

type Publisher interface {
	Publish(message string) *Error
}