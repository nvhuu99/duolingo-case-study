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
	Subscribe(topic string) error
	UnSubscribe(topic string) error
	Listening(ctx context.Context, topic string, closure func(context.Context, string)) error

	SetMainTopic(topic string)
	SubscribeMainTopic() error
	UnSubscribeMainTopic() error
	ListeningMainTopic(ctx context.Context, closure func(context.Context, string)) error
}
