package rabbitmq

import (
	"context"
	mq "duolingo/lib/message_queue"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQTopology struct {
	topics  map[string]*Topic
	manager mq.Manager
	opts    *mq.TopologyOptions
	id      string
	name    string

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex

	errChan chan error
	isReady bool
}

func NewRabbitMQTopology(name string, ctx context.Context) *RabbitMQTopology {
	client := RabbitMQTopology{}
	client.ctx, client.cancel = context.WithCancel(ctx)
	client.topics = make(map[string]*Topic)
	client.opts = mq.DefaultTopologyOptions()
	client.isReady = false
	client.name = name

	return &client
}

func (client *RabbitMQTopology) WithOptions(opts *mq.TopologyOptions) *mq.TopologyOptions {
	if opts == nil {
		client.opts = mq.DefaultTopologyOptions()
	} else {
		client.opts = opts
	}
	return client.opts
}

func (client *RabbitMQTopology) OnConnectionFailure(err error) {
}

func (client *RabbitMQTopology) OnClientFatalError(err error) {
	// client.terminate(err)
}

func (client *RabbitMQTopology) OnReConnected() {
	client.Declare()
}

func (client *RabbitMQTopology) UseManager(manager mq.Manager) {
	client.id = manager.RegisterClient(client.name, client)
	client.manager = manager
}

func (client *RabbitMQTopology) Topic(name string) *Topic {
	if _, found := client.topics[name]; !found {
		client.topics[name] = &Topic{
			name:   name,
			queues: make(map[string]*Queue),
		}
	}

	return client.topics[name]
}

func (client *RabbitMQTopology) NotifyError(ch chan error) chan error {
	client.errChan = ch
	return ch
}

func (client *RabbitMQTopology) IsReady() bool {
	client.mu.RLock()
	defer client.mu.RUnlock()
	return client.isReady
}

func (client *RabbitMQTopology) Declare() error {
	client.mu.Lock()
	client.isReady = false
	client.mu.Unlock()

	var declareErr error
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
				bindings = append(bindings, [3]string{q, p, t})
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
			return fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_DECLARE_TIMEOUT_EXCEED], declareErr)
		default:
		}

		if !firstTry {
			time.Sleep(client.opts.GraceTimeOut)
		}
		firstTry = false

		ch = client.getChannel()
		if ch == nil {
			declareErr = errors.New(mq.ErrMessages[mq.ERR_CONNECTION_FAILURE])
			continue
		}
		if declareErr = client.declareTopics(ch, topics); declareErr != nil {
			continue
		}
		if declareErr = client.declareQueues(ch, queues); declareErr != nil {
			continue
		}
		if declareErr = client.declareBindings(ch, bindings); declareErr != nil {
			continue
		}
		if declareErr = client.declareQos(ch); declareErr != nil {
			continue
		}

		break
	}

	client.mu.Lock()
	client.isReady = true
	client.mu.Unlock()

	return nil
}

func (client *RabbitMQTopology) CleanUp() error {
	if !client.IsReady() {
		return nil
	}

	ch := client.getChannel()
	if ch == nil {
		return errors.New(mq.ErrMessages[mq.ERR_CONNECTION_FAILURE])
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
			return fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_TOPOLOGY_FAILURE], err)
		}
	}

	for t := range topics {
		if err := ch.ExchangeDelete(t, false, false); err != nil {
			return fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_TOPOLOGY_FAILURE], err)
		}
	}

	return nil
}

func (client *RabbitMQTopology) declareTopics(ch *amqp.Channel, topics []string) error {
	for i := 0; i < len(topics); i++ {
		err := ch.ExchangeDeclare(
			topics[i],
			"direct",
			true,  // durable
			false, // delete when unused
			false, // internal
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_TOPIC_DECLARE_FAILURE], err)
		}
	}

	return nil
}

func (client *RabbitMQTopology) declareQueues(ch *amqp.Channel, queues map[string]bool) error {
	for q := range queues {
		_, err := ch.QueueDeclare(
			q,     // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_QUEUE_DECLARE_FAILURE], err)
		}
		if client.opts.QueuesPurged {
			_, err = ch.QueuePurge(q, false)
			if err != nil {
				return fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_QUEUE_DECLARE_FAILURE], err)
			}
		}
	}

	return nil
}

func (client *RabbitMQTopology) declareBindings(ch *amqp.Channel, bindings [][3]string) error {
	for i := 0; i < len(bindings); i++ {
		binding := bindings[i]
		err := ch.QueueBind(binding[0], binding[1], binding[2], false, nil)
		if err != nil {
			return fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_BINDING_DECLARE_FAILURE], err)
		}
	}

	return nil
}

func (client *RabbitMQTopology) declareQos(ch *amqp.Channel) error {
	err := ch.Qos(
		1,    // Prefetch count: One message at a time
		0,    // No size limit for message content
		true, // Apply all channels
	)
	if err != nil {
		return fmt.Errorf("%v - %w", mq.ErrMessages[mq.ERR_DECLARE_FAILURE], err)
	}

	return nil
}

// func (client *RabbitMQTopology) notifyErr(err error) {
// 	if client.errChan != nil {
// 		client.errChan <- err
// 	}
// }

// func (client *RabbitMQTopology) terminate(err error) {
// 	client.notifyErr(err)
// 	client.cancel()
// 	client.manager.UnRegisterClient(client.id)
// }

func (client *RabbitMQTopology) getChannel() *amqp.Channel {
	ch, _ := client.manager.GetClientConnection(client.id)
	channel, ok := ch.(*amqp.Channel)
	if !ok || channel == nil {
		return nil
	}

	return channel
}
