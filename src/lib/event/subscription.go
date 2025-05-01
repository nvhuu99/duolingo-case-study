package event

type strTopicSubscription struct {
	wait bool
	topic string
}

type regexTopicSubscription struct {
	wait bool
	regex *regexPattern
}