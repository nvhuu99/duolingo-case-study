package rabbitmq

import (
	"context"
	mq "duolingo/lib/message-queue"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQTopology struct {
	topics	map[string]*Topic
	manager	mq.Manager
	opts    *mq.TopologyOptions
	id		string
	
	ctx		context.Context
	cancel	context.CancelFunc
	mu		sync.RWMutex

	errChan		chan *mq.Error
	isReady		bool
}

func NewRabbitMQTopology(ctx context.Context, opts *mq.TopologyOptions) *RabbitMQTopology {
	client := RabbitMQTopology{}
	client.ctx, client.cancel = context.WithCancel(ctx)
	client.topics = make(map[string]*Topic)
	if opts == nil {
		opts = mq.DefaultTopologyOptions()
	}
	client.opts = opts
	client.isReady = false

	return &client
}

func (client *RabbitMQTopology) OnConnectionFailure(err *mq.Error) {
}

func (client *RabbitMQTopology) OnClientFatalError(err *mq.Error) {
	client.terminate(err)
}

func (client *RabbitMQTopology) OnReConnected() {
	client.Declare()
}

func (client *RabbitMQTopology) UseManager(manager mq.Manager) {
	client.id = manager.RegisterClient(client)
	client.manager = manager
}

func (client *RabbitMQTopology) Topic(name string) *Topic {
	if _, found := client.topics[name]; !found {
		client.topics[name] = &Topic{ 
			name: name,
			queues: make(map[string]*Queue), 
		}
	}

	return client.topics[name]
}

func (client *RabbitMQTopology) NotifyError(ch chan *mq.Error) chan *mq.Error {
	client.errChan = ch
	return ch
}

func (client *RabbitMQTopology) IsReady() bool {
	client.mu.RLock()
	defer client.mu.RUnlock()
	return client.isReady
}

func (client *RabbitMQTopology) Declare() *mq.Error {
	client.mu.Lock()
	client.isReady = false
	client.mu.Unlock()

	var declareErr *mq.Error
	var ch *amqp.Channel
	var topics []string
	var queues map[string]bool = make(map[string]bool)
	var bindings [][3]string

	for t, topic := range client.topics {
		topics = append(topics, t)
		for q, queue := range topic.queues {
			if len(queue.bindings) == 0 {
				continue
			}
			queues[q] = true
			for p := range queue.bindings {
				bindings = append(bindings, [3]string { q, p, t })
			}
		}
	}
	
	declareDeadline := time.After(client.opts.DeclareTimeOut)
	firstTry := true
	for {
		select {
		case <-client.ctx.Done():
			return declareErr
		case <-declareDeadline:
			return mq.NewError(mq.DeclareTimeOutExceed, declareErr, "", "", "")
		default:
		}

		if ! firstTry {
			time.Sleep(client.opts.GraceTimeOut)
		}
		firstTry = false

		ch = client.getChannel()
		if ch == nil {
			declareErr = mq.NewError(mq.ConnectionFailure, nil, "", "", "")
			continue
		}
		if declareErr = declareTopics(ch, topics); declareErr != nil {
			continue
		}
		if declareErr = declareQueues(ch, queues); declareErr != nil {
			continue
		}
		if declareErr = declareBindings(ch, bindings); declareErr != nil {
			continue
		}
		if declareErr = declareQos(ch); declareErr != nil {
			continue
		}

		break
	}

	client.mu.Lock()
	client.isReady = true
	client.mu.Unlock()

	return nil
}

func (client *RabbitMQTopology) CleanUp() *mq.Error {
	if ! client.IsReady() {
		return nil
	}

	ch := client.getChannel()
	if ch == nil {
		return mq.NewError(mq.ConnectionFailure, nil, "", "", "") 
	}

	client.mu.Lock()
	client.isReady = false
	client.mu.Unlock()

	topics := make(map[string]bool)
	queues := make(map[string]bool)
	for t, topic := range client.topics {
		topics[t] = true
		for q := range topic.queues {
			queues[q] = true
		}
	}

	for q := range queues {
		if _, err := ch.QueueDelete(q, false, false, false); err != nil {
			return mq.NewError(mq.TopologyFailure, err, "", q, "")
		}
	}

	for t := range topics {
		if err := ch.ExchangeDelete(t, false, false); err != nil {
			return mq.NewError(mq.TopologyFailure, err, t, "", "")
		}
	}

	return nil
}

func (client *RabbitMQTopology) notifyErr(err *mq.Error) {
	if client.errChan != nil {
		client.errChan <- err
	}
}

func (client *RabbitMQTopology) terminate(err *mq.Error) {
	client.notifyErr(err)
	client.cancel()
	client.manager.UnRegisterClient(client.id)
}

func (client *RabbitMQTopology) getChannel() *amqp.Channel {
	ch, _ := client.manager.GetClientConnection(client.id)
	channel, ok := ch.(*amqp.Channel)
	if !ok || channel == nil {
		return nil
	}

	return channel
}

func declareTopics(ch *amqp.Channel, topics []string) *mq.Error {
	for i := 0; i < len(topics); i++ {
		err := ch.ExchangeDeclare(
			topics[i],
			"direct",
			true,              // durable
			false,             // delete when unused
			false,             // internal
			false,             // no-wait
			nil,               // arguments
		)
		if err != nil {
			return mq.NewError(mq.TopicDeclareFailure, err, topics[i], "", "")
		}
	}

	return nil
}

func declareQueues(ch *amqp.Channel, queues map[string]bool) *mq.Error {
	for q := range queues {
		_, err := ch.QueueDeclare(
			q,			// name
			true,		// durable
			false,		// delete when unused
			false,		// exclusive
			false,		// no-wait
			nil,		// arguments
		)
		if err != nil {
			return mq.NewError(mq.QueueDeclareFailure, err, q, "", "")
		}
	}

	return nil
}

func declareBindings(ch *amqp.Channel, bindings [][3]string) *mq.Error {
	for i := 0; i < len(bindings); i++ {
		binding := bindings[i]
		err := ch.QueueBind(binding[0], binding[1], binding[2], false, nil)
		if err != nil {
			return mq.NewError(mq.BindingDeclareFailure, err, binding[2], binding[1], binding[0])
		}
	}

	return nil
}

func declareQos(ch *amqp.Channel) *mq.Error {
	err := ch.Qos(
		1,     // Prefetch count: One message at a time
		0,     // No size limit for message content
		true,  // Apply all channels
	)
	if err != nil {
		return mq.NewError(mq.DeclareFailure, err, "", "", "")
	}

	return nil
}

