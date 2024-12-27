package topicQueue

type DistributeMethod string

const (
	QueueFanout DistributeMethod = "queueFanout"
	QueueDispatch DistributeMethod = "queueDispatch"
)

type MessageQueueTopic interface {
	UseConnectionString(uri string)
	UseConnection(host string, port string, user string, pwd string)

	SetTopic(topic string)
	SetNumberOfQueue(total int)
	SetQueueWorkerLimit(total int)
	SetDistributeMethod(method DistributeMethod)

	Publish() error
	Shutdown()

	GetFirstAvailableQueue() (string, error)
	GetWorkerQueue(worker string) (string, error)
	UseWorker(worker string, queue string) error
}