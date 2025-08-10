package events

import "github.com/google/uuid"

type Subscriber interface {
	Notify(event *Event)
	GetId() string
}

type BaseEventSubscriber struct {
	subscriberId string
}

func NewBaseEventSubscriber() *BaseEventSubscriber {
	return &BaseEventSubscriber{uuid.NewString()}
}

func (base *BaseEventSubscriber) Notify(event *Event) {
}

func (base *BaseEventSubscriber) GetId() string {
	return base.subscriberId
}
