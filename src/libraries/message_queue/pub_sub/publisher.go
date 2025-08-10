package pub_sub

import (
	"context"
	"errors"
)

var (
	ErrPublisherMainTopicNotSet = errors.New("publisher main topic is not set")
)

type Publisher interface {
	DeclareTopic(ctx context.Context, topic string) error
	RemoveTopic(ctx context.Context, topic string) error
	Notify(ctx context.Context, topic string, message string) error

	SetMainTopic(topic string)
	DeclareMainTopic(ctx context.Context) error
	RemoveMainTopic(ctx context.Context) error
	NotifyMainTopic(ctx context.Context, message string) error
}
