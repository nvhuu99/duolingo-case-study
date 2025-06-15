package pub_sub

import "context"

type Subscriber interface {
	GetChannel() string
	Consuming(ctx context.Context, topic string, closure func(string) ConsumeAction) error
}
