package pub_sub

type Publisher interface {
	DeclareTopic(topic string) error
	RemoveTopic(topic string) error
	Notify(topic string, message string) error

	SetMainTopic(topic string)
	DeclareMainTopic() error
	RemoveMainTopic() error
	NotifyMainTopic(message string) error
}
