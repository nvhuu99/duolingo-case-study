package pub_sub

type Publisher interface {
	AddSubscriber(topic string, subscriber Subscriber) error
	RemoveSubscriber(subscriber Subscriber) error
	Notify(topic string, message string) error
}
