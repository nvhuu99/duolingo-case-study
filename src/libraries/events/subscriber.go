package events

import "github.com/google/uuid"

type Subscriber interface {
	Notify(event *Event)
	GetId() string
}

type BaseEventSubscriber struct {
	subscriberId string
	notifyFunc   func(*Event)
}

func NewBaseEventSubscriber() *BaseEventSubscriber {
	return &BaseEventSubscriber{
		subscriberId: uuid.NewString(),
	}
}

func (base *BaseEventSubscriber) Notify(event *Event) {
	base.notifyFunc(event)
}

func (base *BaseEventSubscriber) GetId() string {
	return base.subscriberId
}
