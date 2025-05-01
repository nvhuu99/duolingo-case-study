package rabbitmq

// Event data types

const (
	EVT_ON_CLIENT_ACTION string = "evt_on_client_action"
	EVT_CONNECTION_FAILURE string = "evt_connection_failure"
	EVT_CLIENT_FATAL_ERR string = "evt_client_fatal_err"
)

type ClientActionEvent struct {
	ClientName   string
	QueueName    string
	Action ClientAction
}

type ConnectionFailureEvent struct {
	Error error
}

type ClientFatalErr struct {
	Id string
	ClientName string
	Error error
}

// Client actions

type ClientAction string

const (
	ConsumerAccept  ClientAction = "consumer_accept"
	ConsumerRequeue ClientAction = "consumer_reject_requeue"
	ConsumerReject  ClientAction = "consumer_reject"
	PublisherPublished ClientAction = "publisher_published"
	TopologyDeclared ClientAction = "topology_declared"
)
