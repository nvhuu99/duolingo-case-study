package pub_sub

import (
	"context"
	"errors"
)

var (
	ErrPublisherMainTopicNotSet = errors.New("publisher main topic is not set")
)

type Publisher interface {
	DeclareTopic(topic string) error
	RemoveTopic(topic string) error
	Notify(ctx context.Context, topic string, message string) error

	SetMainTopic(topic string)
	DeclareMainTopic() error
	RemoveMainTopic() error
	NotifyMainTopic(ctx context.Context, message string) error
}
