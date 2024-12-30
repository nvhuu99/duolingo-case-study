package messagequeue

type MessagePublisher interface {
	SetTopicInfo(info TopicInfo)
	Publish(message string) error
}
