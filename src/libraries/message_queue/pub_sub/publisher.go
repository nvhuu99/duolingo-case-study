package pub_sub

import "errors"

var (
	ErrPublisherMainTopicNotSet = errors.New("publisher main topic is not set")
)

type Publisher interface {
	DeclareTopic(topic string) error
	RemoveTopic(topic string) error
	Notify(topic string, message string) error

	SetMainTopic(topic string)
	DeclareMainTopic() error
	RemoveMainTopic() error
	NotifyMainTopic(message string) error
}
