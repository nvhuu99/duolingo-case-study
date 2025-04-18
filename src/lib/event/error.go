package event

const (
	ERR_SUBSCRIBER_TOPIC_EXISTED = 501
	ERR_SUBSCRIBER_NOT_EXIST     = 502
	ERR_EMPTY_PATTERN            = 503
	ERR_SUBCRIBER_ID_EMPTY       = 504
)

var ErrorMessages = map[int]string{
	ERR_SUBSCRIBER_TOPIC_EXISTED: "subcriber topic \"%v\"is already registered",
	ERR_SUBSCRIBER_NOT_EXIST:     "subcriber is not registered",
	ERR_EMPTY_PATTERN:            "can not register with an empty pattern",
	ERR_SUBCRIBER_ID_EMPTY:       "can not register with an empty subcriber id",
}
