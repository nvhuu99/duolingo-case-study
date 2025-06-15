package pub_sub

type ConsumeAction string

const (
	ActionRequeue ConsumeAction = "requeue_message"
	ActionAccept  ConsumeAction = "accept_message"
	ActionReject  ConsumeAction = "reject_message"
)
