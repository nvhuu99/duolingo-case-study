package events

import "log"

type Subscriber interface {
	Notified(event *Event)
}

type SubscriberImp struct {
	Name string
}

func (sub *SubscriberImp) Notified(event *Event) {
	log.Println(sub.Name, "received event", event.name, event.startedAt, event.endedAt)
}
