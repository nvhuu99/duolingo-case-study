package pub_sub

import "context"

type Subscriber interface {
	Subscribe(topic string) error
	UnSubscribe(topic string) error
	Consuming(ctx context.Context, topic string, closure func(string) ConsumeAction) error
}
