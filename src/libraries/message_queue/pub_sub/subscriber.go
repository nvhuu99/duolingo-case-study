package pub_sub

import (
	"context"
	"errors"
)

var (
	ErrSubscriberMainTopicNotSet    = errors.New("subscriber's main topic not set")
	ErrSubscriberTopicNotSubscribed = errors.New("failed to listen, the topic is not subscribed")
)

type Subscriber interface {
	Subscribe(ctx context.Context, topic string) error
	UnSubscribe(ctx context.Context, topic string) error
	Listening(ctx context.Context, topic string, closure func(context.Context, string) error) error

	SetMainTopic(topic string)
	SubscribeMainTopic(ctx context.Context) error
	UnSubscribeMainTopic(ctx context.Context) error
	ListeningMainTopic(ctx context.Context, closure func(context.Context, string) error) error
}
