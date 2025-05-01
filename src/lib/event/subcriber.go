package event

type Subcriber interface {
	SubscriberId() string
	Notified(string, any)
}
