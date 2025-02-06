package messagequeue

type ConsumerAction string

const (
	ConsumerAccept				= "ack"
	ConsumerRejectAndRequeue	= "nack_requeue"
	ConsumerRejectAndDrop		= "nack_drop"
)
