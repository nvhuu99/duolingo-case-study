package messagequeue

type ConsumerAction string

const (
	ConsumerAccept  ConsumerAction = "ack"
	ConsumerRequeue ConsumerAction = "nack_requeue"
	ConsumerReject  ConsumerAction = "nack_reject"
)
