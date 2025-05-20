package event

type Subscriber interface {
	SubscriberId() string
	Notified(string, any)
}
