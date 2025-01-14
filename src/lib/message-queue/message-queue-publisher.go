package messagequeue

type MessagePublisher interface {
	SetTopicInfo(info *TopicInfo)
	Connect() error
	Disconnect()
	Publish(message string) error
}
