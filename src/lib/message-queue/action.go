package messagequeue

type ConsumerAction string

const (
	ConsumerAccept  = "ack"
	ConsumerRequeue = "nack_requeue"
	ConsumerReject  = "nack_reject"
)
