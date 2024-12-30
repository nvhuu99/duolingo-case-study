package messagequeue

import "fmt"

type MessageError struct {
	ErrorMessage string
	Message      string
	Topic        string
	Pattern      string
}

func (e *MessageError) Error() string {
	return fmt.Sprintf(`error sending message: error: "%s", topic: "%s", pattern: "%s", message: "%s"`,
		e.ErrorMessage, e.Topic, e.Pattern, e.Message)
}

func NewMessageError(err string, msg string, topic string, pattern string) error {
	return &MessageError{
		ErrorMessage: err,
		Message:      msg,
		Topic:        topic,
		Pattern:      pattern,
	}
}
