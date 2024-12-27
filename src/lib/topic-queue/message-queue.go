package topicQueue

type MessageQueue interface {
	UseConnectionString(uri string)
	UseConnection(host string, port string, user string, pwd string)
	SetPublishRoute(topic string, pattern string)
	SetConsumeQueue(queue string)

	Open() error
	Close()
	Publish(message string) error
	Consume(hanlder func(string) bool) error

	Error() error
}