package event

type Publisher interface {
	Subscribe(wait bool, topic string, sub Subcriber) error
	SubscribeRegex(wait bool, reg string, sub Subcriber) error
	UnSubscribe(topic string, sub Subcriber) error
	Notify(topic string, data any)
}
