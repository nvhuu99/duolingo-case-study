package rabbitmq

import (
	"context"
	tq "duolingo/lib/topic-queue" // message queue service
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type topicStatus string

const (
	statusNotPublished = "notPublished"
	statusPublished = "published"
	statusShutdown = "shutdown"

	errTopicClosed    = "topic has not yet been published"
	errTopicNotSet    = "topic is not set"
	errWorkerQueueNotExists = "the worker is not registered"
	errQueueWorkerFull = "no available queue to register new worker"
)

type RabbitMQTopic struct {
	uri string
	name string
	numQueue int
	numWorkerPerQueue int
	status topicStatus
	method tq.DistributeMethod

	queues []string
	queueWorkers map[string][]string

	conn *amqp.Connection
	ch   *amqp.Channel

	ctx context.Context
	mu  sync.Mutex
}

func NewTopic(ctx context.Context) tq.MessageQueueTopic {
	topic := RabbitMQTopic{}
	topic.ctx = ctx
	topic.numQueue = 1
	topic.numWorkerPerQueue = 1
	topic.method = tq.QueueDispatch
	topic.status = statusNotPublished
	topic.queues = make([]string, topic.numQueue)
	topic.queueWorkers = make(map[string][]string, topic.numQueue)

	return &topic
}

// Sets the URI for the RabbitMQ connection.
func (topic *RabbitMQTopic) UseConnectionString(uri string) {
	if topic.status != statusPublished {
		topic.uri = uri
	}
}

// Sets the connection parameters (host, port, user, password) and generates the URI.
func (topic *RabbitMQTopic) UseConnection(host string, port string, user string, pwd string) {
	if topic.status != statusPublished {
		pwd = url.QueryEscape(pwd)
		topic.uri = fmt.Sprintf("amqp://%v:%v@%v:%v/", user, pwd, host, port)
	}
}

func (topic *RabbitMQTopic) SetTopic(name string) {
	if topic.status != statusPublished {
		topic.name = name
	}
}

func (topic *RabbitMQTopic) SetNumberOfQueue(total int) {
	if topic.status != statusPublished {
		if total > 0 {
			topic.numQueue = total
		}
	}
}

func (topic *RabbitMQTopic) SetQueueWorkerLimit(total int) {
	if topic.status != statusPublished {
		if total > 0 {
			topic.numWorkerPerQueue = total
		}
	}
}

func (topic *RabbitMQTopic) SetDistributeMethod(method tq.DistributeMethod) {
	if topic.status != statusPublished {
		topic.method = method
	}
}


// Be warning that before Publish(), will first try to remove the topic and related queues
// to provide a clean work state 
func (topic *RabbitMQTopic) Publish() error {
	if topic.status == statusPublished {
		return nil
	}

	if topic.name == "" {
		return errors.New(errTopicNotSet)
	}

	// after open, must delete old exchange and queues
	err := topic.open()
	topic.deleteQueues()
	topic.deleteExchange()
	if err != nil {
		return err
	}

	err = topic.createExchange() 
	if err != nil {
		return err
	}

	err = topic.createQueues()
	if err != nil {
		return err
	}

	err = topic.bindQueues()
	if err != nil {
		return err
	}

	topic.status = statusPublished
	for _, queue := range topic.queues {
		topic.queueWorkers[queue] = make([]string, 0)
	}

	return nil
}

func (topic *RabbitMQTopic) Shutdown() {
	topic.mu.Lock()
	topic.status = statusShutdown
	topic.mu.Unlock()

	topic.deleteQueues()
	topic.deleteExchange()

	topic.ch.Close()
	topic.conn.Close()
}

func (topic *RabbitMQTopic) GetFirstAvailableQueue() (string, error) {
	topic.mu.Lock()
	defer topic.mu.Unlock()

	if topic.status != statusPublished {
		return "", errors.New(errTopicClosed)
	}
	
	for queue, workers := range topic.queueWorkers {
		if len(workers) < topic.numWorkerPerQueue {
			return queue, nil
		}
	}

	return "", errors.New(errQueueWorkerFull)
}

func (topic *RabbitMQTopic) GetWorkerQueue(worker string) (string, error) {
	topic.mu.Lock()
	defer topic.mu.Unlock()
	
	for q, workers := range topic.queueWorkers {
		for _, w := range workers {
			if w == worker {
				return q, nil
			}
		}
	}

	return "", errors.New(errWorkerQueueNotExists)
}

func (topic *RabbitMQTopic) UseWorker(worker string, queue string) error {
	if _, err := topic.GetFirstAvailableQueue(); err != nil {
		return err
	}

	topic.mu.Lock()
	defer topic.mu.Unlock()

	topic.queueWorkers[queue] = append(topic.queueWorkers[queue], worker)

	return nil
}

func (topic *RabbitMQTopic) open() error {
	if topic.status == statusPublished {
		return nil
	}
	conn, err := amqp.Dial(topic.uri)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	topic.conn = conn
	topic.ch = ch

	return nil
}

func (topic *RabbitMQTopic) createExchange() error {
	err := topic.ch.ExchangeDeclare(
		topic.name,
		"direct",
		true,              // durable
		false,             // delete when unused
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	return err
}

func (topic *RabbitMQTopic) createQueues() error {
	for i := 1; i <= topic.numQueue; i++ {
		qName := topic.getQueueNameByIndex(i)
		
		_, err := topic.ch.QueueDeclare(
			qName,        // name
			true,         // durable
			false,        // delete when unused
			false,        // exclusive
			false,        // no-wait
			nil,          // arguments
		)

		topic.queues[i] = qName 

		if err != nil {
			return err
		}
	}

	return nil
}

func (topic *RabbitMQTopic) bindQueues() error {
	for i := 1; i <= topic.numQueue; i++ {
		qName := topic.queues[i]

		var routingKey string
		if topic.method == tq.QueueFanout {
			routingKey = ""
		}
		if topic.method == tq.QueueDispatch {
			routingKey = qName
		}

		err := topic.ch.QueueBind(qName, routingKey, topic.name, false, nil)

		if err != nil {
			return err
		}
	}

	return nil
}

func (topic *RabbitMQTopic) getQueueNameByIndex(i int) string {
	return fmt.Sprintf("%v.queue_%v", topic.name, i)
}

func (topic *RabbitMQTopic) deleteExchange() {
	topic.ch.ExchangeDelete(topic.name, false, false)
	log.Printf("topic shutdown: topic %v deleted\n", topic.name)
}

func (topic *RabbitMQTopic) deleteQueues() {
	for i := 1; i <= topic.numQueue; i++ {
		topic.ch.QueueDelete(topic.queues[i], false, false, false)
		log.Printf("topic shutdown: queue %v deleted\n", topic.queues[i])
	}
}
