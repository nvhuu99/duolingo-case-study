package rabbitmq

/* ConsumeAction */

type ConsumeAction string

const (
	ActionRequeue ConsumeAction = "requeue_message"
	ActionAccept  ConsumeAction = "accept_message"
	ActionReject  ConsumeAction = "reject_message"
)

/* QueueOptions */

type QueueOptions struct {
	name       string
	durable    bool
	autoDelete bool
	exclusive  bool
}

func DefaultQueueOpts(name string) *QueueOptions {
	return &QueueOptions{
		name:       name,
		durable:    false,
		autoDelete: false,
		exclusive:  false,
	}
}

func (opts *QueueOptions) IsPersistent() *QueueOptions {
	opts.durable = true
	opts.autoDelete = false
	return opts
}

func (opts *QueueOptions) IsNonPersistent() *QueueOptions {
	opts.durable = false
	opts.autoDelete = true
	return opts
}

func (opts *QueueOptions) IsExclusive() *QueueOptions {
	opts.exclusive = true
	return opts
}

/* ExchangeOptions */

type ExchangeType string

const (
	FanoutExchange ExchangeType = "fanout"
	DirectExchange ExchangeType = "direct"
	TopicExchange  ExchangeType = "topic"
)

type ExchangeOptions struct {
	name       string
	kind       ExchangeType
	durable    bool
	autoDelete bool
}

func DefaultExchangeOpts(name string) *ExchangeOptions {
	return &ExchangeOptions{
		name:       name,
		kind:       FanoutExchange,
		durable:    false,
		autoDelete: false,
	}
}

func (opts *ExchangeOptions) IsType(value ExchangeType) *ExchangeOptions {
	opts.kind = value
	return opts
}

func (opts *ExchangeOptions) IsPersistent() *ExchangeOptions {
	opts.durable = true
	opts.autoDelete = false
	return opts
}

func (opts *ExchangeOptions) IsNonPersistent() *ExchangeOptions {
	opts.durable = false
	opts.autoDelete = true
	return opts
}

/* QueueBindings */

type binding struct {
	routingKey string
	exchange   string
}

type QueueBindings struct {
	queue    string
	bindings []*binding
}

func NewQueueBinding(queue string) *QueueBindings {
	return &QueueBindings{queue: queue}
}

func (qb *QueueBindings) Add(routingKey string, exchange string) *QueueBindings {
	qb.bindings = append(qb.bindings, &binding{
		routingKey: routingKey,
		exchange:   exchange,
	})
	return qb
}
