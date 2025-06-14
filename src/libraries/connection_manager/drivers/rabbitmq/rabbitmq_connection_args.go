package rabbitmq

import (
	"duolingo/libraries/connection_manager"
	"time"
)

type RabbitMQConnectionArgs struct {
	*connection_manager.BaseConnectionArgs

	uri            string
	host           string
	port           string
	user           string
	password       string
	declareTimeout time.Duration
	heartbeat      time.Duration

	prefetchCount uint8
	prefetchLimit uint
}

func DefaultRabbitMQConnectionArgs() *RabbitMQConnectionArgs {
	baseArgs := connection_manager.DefaultConnectionArgs()
	redisArgs := &RabbitMQConnectionArgs{
		BaseConnectionArgs: baseArgs,
		host:               "127.0.0.1",
		port:               "5672",
		user:               "",
		password:           "",
		declareTimeout:     15 * time.Second,
		heartbeat:          20 * time.Second,
		prefetchCount:      1,
		prefetchLimit:      0, // no size limit for message content
	}
	return redisArgs
}

func (r *RabbitMQConnectionArgs) GetURI() string {
	return r.uri
}

func (r *RabbitMQConnectionArgs) SetURI(uri string) *RabbitMQConnectionArgs {
	r.uri = uri
	return r
}

func (r *RabbitMQConnectionArgs) GetHost() string {
	return r.host
}

func (r *RabbitMQConnectionArgs) SetHost(host string) *RabbitMQConnectionArgs {
	r.host = host
	return r
}

func (r *RabbitMQConnectionArgs) GetPort() string {
	return r.port
}

func (r *RabbitMQConnectionArgs) SetPort(port string) *RabbitMQConnectionArgs {
	r.port = port
	return r
}

func (r *RabbitMQConnectionArgs) GetUser() string {
	return r.user
}

func (r *RabbitMQConnectionArgs) SetUser(user string) *RabbitMQConnectionArgs {
	r.user = user
	return r
}

func (r *RabbitMQConnectionArgs) GetPassword() string {
	return r.password
}

func (r *RabbitMQConnectionArgs) SetPassword(password string) *RabbitMQConnectionArgs {
	r.password = password
	return r
}

func (r *RabbitMQConnectionArgs) GetHeartbeat() time.Duration {
	return r.heartbeat
}

func (r *RabbitMQConnectionArgs) SetHeartbeat(value time.Duration) *RabbitMQConnectionArgs {
	r.heartbeat = value
	return r
}

func (r *RabbitMQConnectionArgs) GetDeclareTimeout() time.Duration {
	return r.declareTimeout
}

func (r *RabbitMQConnectionArgs) SetDeclareTimeout(value time.Duration) *RabbitMQConnectionArgs {
	r.declareTimeout = value
	return r
}

func (r *RabbitMQConnectionArgs) GetPrefetchCount() uint8 {
	return r.prefetchCount
}

func (r *RabbitMQConnectionArgs) SetPrefetchCount(value uint8) *RabbitMQConnectionArgs {
	r.prefetchCount = value
	return r
}

func (r *RabbitMQConnectionArgs) GetPrefetchLimit() uint {
	return r.prefetchLimit
}

func (r *RabbitMQConnectionArgs) SetPrefetchLimit(value uint) *RabbitMQConnectionArgs {
	r.prefetchLimit = value
	return r
}
