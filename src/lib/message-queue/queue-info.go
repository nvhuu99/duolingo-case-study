package messagequeue

type QueueInfo struct {
	ConnectionString string		`json:"connectionString"`
	QueueName        string		`json:"queueName"`
	ConsumerLimit    int		`json:"consumerLimit"`
	TotalConsumer    int		`json:"totalConsumer"`
}