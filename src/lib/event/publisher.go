package event

type Publisher interface {
	Subscribe(wait bool, topic string, sub Subscriber) error
	SubscribeRegex(wait bool, reg string, sub Subscriber) error
	UnSubscribe(topic string, sub Subscriber) error
	Notify(topic string, data any)
}
