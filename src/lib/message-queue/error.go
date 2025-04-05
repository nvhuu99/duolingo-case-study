package messagequeue

const (
	ERR_CONNECTION_FAILURE      = 501
	ERR_CONNECTION_TIMEOUT      = 502
	ERR_PUBLISH_FAILURE         = 503
	ERR_PUBLISH_CONFIRM_FAILURE = 504
	ERR_PUBLISH_NACK            = 505
	ERR_PUBLISH_TIMEOUT_EXCEED  = 506
	ERR_DECLARE_FAILURE         = 507
	ERR_DECLARE_TIMEOUT_EXCEED  = 508
	ERR_TOPIC_DECLARE_FAILURE   = 509
	ERR_QUEUE_DECLARE_FAILURE   = 510
	ERR_BINDING_DECLARE_FAILURE = 511
	ERR_TOPOLOGY_FAILURE        = 512
	ERR_CLIENT_FATAL_ERROR      = 513
	ERR_MANAGER_CONFIG_MISSING  = 514
)

var ErrMessages = map[int]string{
	ERR_CONNECTION_FAILURE:      "501 - connection failure",
	ERR_CONNECTION_TIMEOUT:      "502 - connection timeout",
	ERR_PUBLISH_FAILURE:         "503 - publish message failure",
	ERR_PUBLISH_CONFIRM_FAILURE: "504 - publish message confirm failure",
	ERR_PUBLISH_NACK:            "505 - publish message not acknowledged (NACK)",
	ERR_PUBLISH_TIMEOUT_EXCEED:  "506 - publish message timeout exceeded",
	ERR_DECLARE_FAILURE:         "507 - declare failure",
	ERR_DECLARE_TIMEOUT_EXCEED:  "508 - declare timeout exceeded",
	ERR_TOPIC_DECLARE_FAILURE:   "509 - topic declare failure",
	ERR_QUEUE_DECLARE_FAILURE:   "510 - queue declare failure",
	ERR_BINDING_DECLARE_FAILURE: "511 - binding declare failure",
	ERR_TOPOLOGY_FAILURE:        "512 - topology operation failure",
	ERR_CLIENT_FATAL_ERROR:      "513 - client operations fatal error",
	ERR_MANAGER_CONFIG_MISSING:  "514 - manager configuration missing",
}
