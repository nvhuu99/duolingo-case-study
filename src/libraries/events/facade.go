package events

import (
	"context"
	"time"
)

func Subscribe(sub Subscriber) {
	GetManager().AddSubsriber(sub)
}

func New(ctx context.Context, name string) (context.Context, *Event) {
	return GetManager().NewEvent(ctx, name)
}

func End(event *Event) {
	GetManager().EndEvent(event, time.Now())
}