package message_queue

type ConsumerAction string

const (
	ConsumerAccept  ConsumerAction = "consumer_accept"
	ConsumerRequeue ConsumerAction = "consumer_reject_requeue"
	ConsumerReject  ConsumerAction = "consumer_reject"
)
