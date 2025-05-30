package rabbitmq

import "time"

// Event data types

const (
	EVT_CLIENT_ACTION_CONSUMED   string = "evt_client_action_consumed"
	EVT_CLIENT_ACTION_PUBLISHED   string = "evt_client_action_published"
	EVT_CONNECTION_FAILURE string = "evt_connection_failure"
	EVT_CLIENT_FATAL_ERR   string = "evt_client_fatal_err"
)

type ConsumeEvent struct {
	ClientName string
	QueueName  string
	Action     ClientAction
}

type PublishEvent struct {
	ClientName string
	QueueName  string
	Action     ClientAction
	Latency time.Duration
}

type ConnectionFailureEvent struct {
	Error error
}

type ClientFatalErr struct {
	Id         string
	ClientName string
	Error      error
}

type ClientAction string

const (
	ConsumerAccept     ClientAction = "consumer_accept"
	ConsumerRequeue    ClientAction = "consumer_reject_requeue"
	ConsumerReject     ClientAction = "consumer_reject"
	PublisherPublished ClientAction = "publisher_published"
	TopologyDeclared   ClientAction = "topology_declared"
)
