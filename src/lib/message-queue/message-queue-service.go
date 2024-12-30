package messagequeue

type MessageQueueService interface {
	UseConnectionString(uri string)
	UseConnection(host string, port string, user string, pwd string)

	SetTopic(topic string)
	SetNumberOfQueue(total int)
	SetQueueConsumerLimit(total int)
	SetDistributeMethod(method DistributeMethod)

	Publish() error
	Shutdown()

	RegisterConsumer(consumer string) error
	GetTopicInfo() TopicInfo
	GetQueueInfo(queue string) (*QueueInfo, error)
}