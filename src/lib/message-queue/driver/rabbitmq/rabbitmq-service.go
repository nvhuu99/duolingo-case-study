package rabbitmq

import (
	"context"
	"duolingo/lib/helper-functions"
	mqp "duolingo/lib/message-queue"
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
	statusPublished    = "published"
	statusShutdown     = "shutdown"

	errTopicClosed            = "topic and queues names have not yet been published"
	errTopicNotSet            = "topic and queues names must be set before publishing"
	errConsumerQueueNotExists = "the consumer is not registered"
	errQueueConsumerFull      = "no available queue to register new consumer"
	errQueueNotExists         = "queue is not exist in this topic"
)

// Singleton
type RabbitMQService struct {
	topic  mqp.TopicInfo
	status topicStatus
	
	numQueue            int
	numConsumerPerQueue int
	consumers           map[string][]string
	
	ctx context.Context
	mu  sync.Mutex
}

func NewMQService(ctx context.Context) *RabbitMQService {
	mq := RabbitMQService{}
	mq.ctx = ctx
	mq.topic = mqp.TopicInfo{}
	mq.status = statusNotPublished
	mq.SetNumberOfQueue(1)
	mq.SetQueueConsumerLimit(1)

	return &mq
}

// Sets the URI for the RabbitMQ connection.
func (mq *RabbitMQService) UseConnectionString(uri string) {
	if mq.status != statusPublished {
		mq.topic.ConnectionString = uri
	}
}

// Sets the connection parameters (host, port, user, password) and generates the URI.
func (mq *RabbitMQService) UseConnection(host string, port string, user string, pwd string) {
	if mq.status != statusPublished {
		pwd = url.QueryEscape(pwd)
		uri := fmt.Sprintf("amqp://%v:%v@%v:%v/", user, pwd, host, port) 
		mq.topic.ConnectionString = uri
	}
}

func (mq *RabbitMQService) SetTopic(name string) {
	if mq.status != statusPublished {
		mq.topic.Name = name
	}
}

func (mq *RabbitMQService) SetNumberOfQueue(total int) {
	if mq.status != statusPublished {
		if total > 0 {
			mq.numQueue = total
			mq.consumers = make(map[string][]string, total)
			mq.topic.Queues = make([]string, total)
			for i := 0; i < mq.numQueue; i++ {
				mq.topic.Queues[i] = mq.getQueueNameByIndex(i + 1)
			}
		}
	}
}

func (mq *RabbitMQService) SetQueueConsumerLimit(total int) {
	if mq.status != statusPublished {
		if total > 0 {
			mq.numConsumerPerQueue = total
		}
	}
}

func (mq *RabbitMQService) SetDistributeMethod(method mqp.DistributeMethod) {
	if mq.status != statusPublished {
		mq.topic.Method = method
	}
}

func (mq *RabbitMQService) Publish() error {
	if mq.status == statusPublished {
		return nil
	}

	if mq.topic.Name == "" {
		return errors.New(errTopicNotSet)
	}

	// must delete old exchange and queues
	mq.deleteQueues()
	mq.deleteExchange()

	err := mq.createExchange() 
	if err != nil {
		return err
	}

	err = mq.createQueues()
	if err != nil {
		return err
	}

	err = mq.bindQueues()
	if err != nil {
		return err
	}

	for _, queue := range mq.topic.Queues {
		mq.consumers[queue] = make([]string, 0)
	}

	mq.status = statusPublished

	return nil
}

func (mq *RabbitMQService) Shutdown() {
	mq.mu.Lock()
	mq.status = statusShutdown
	mq.mu.Unlock()

	mq.deleteQueues()
	mq.deleteExchange()
}

func (mq *RabbitMQService) RegisterConsumer(consumer string) (*mqp.QueueInfo, error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.status != statusPublished {
		return nil, errors.New(errTopicClosed)
	}
	
	for queue, consumers := range mq.consumers {
		if len(consumers) < mq.numConsumerPerQueue {
			mq.consumers[queue] = append(mq.consumers[queue], consumer)
			info, _ := mq.GetQueueInfo(queue)
			return info, nil
		}
	}

	return nil, errors.New(errQueueConsumerFull)
}

func (mq *RabbitMQService) GetTopicInfo() mqp.TopicInfo {
	return mqp.TopicInfo {
		ConnectionString: mq.topic.ConnectionString,
		Name: mq.topic.Name,
		Queues: mq.topic.Queues,
		Method: mq.topic.Method,
	}
}

func (mq *RabbitMQService) GetQueueInfo(queue string) (*mqp.QueueInfo, error) {
	if ! helper.InArray(queue, mq.topic.Queues) {
		return nil, errors.New(errQueueNotExists)
	}

	info := mqp.QueueInfo {
		ConnectionString: mq.topic.ConnectionString,
		QueueName: queue,
		ConsumerLimit: mq.numConsumerPerQueue,
		TotalConsumer: len(mq.consumers[queue]),
	}

	return &info, nil
}

func (mq *RabbitMQService) connect() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(mq.topic.ConnectionString)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

func (mq *RabbitMQService) createExchange() error {
	conn, ch, err := mq.connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	err = ch.ExchangeDeclare(
		mq.topic.Name,
		"direct",
		true,              // durable
		false,             // delete when unused
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	return err
}

func (mq *RabbitMQService) createQueues() error {
	conn, ch, err := mq.connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	for i := 0; i < mq.numQueue; i++ {
		_, err := ch.QueueDeclare(
			mq.topic.Queues[i], // name
			true,         // durable
			false,        // delete when unused
			false,        // exclusive
			false,        // no-wait
			nil,          // arguments
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (mq *RabbitMQService) bindQueues() error {
	conn, ch, err := mq.connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	for i := 0; i < mq.numQueue; i++ {
		qName := mq.topic.Queues[i]

		routingKey, err := mq.topic.Pattern(qName)
		if err != nil {
			return err
		}
		err = ch.QueueBind(qName, routingKey, mq.topic.Name, false, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mq *RabbitMQService) getQueueNameByIndex(i int) string {
	return fmt.Sprintf("%v.queue_%v", mq.topic.Name, i)
}

func (mq *RabbitMQService) deleteExchange() {
	conn, ch, err := mq.connect()
	if err != nil {
		return
	}
	defer conn.Close()
	defer ch.Close()

	err = ch.ExchangeDelete(mq.topic.Name, false, false)
	if err == nil {
		log.Printf("topic %v deleted\n", mq.topic.Name)
	}
}

func (mq *RabbitMQService) deleteQueues() {
	conn, ch, err := mq.connect()
	if err != nil {
		return
	}
	defer conn.Close()
	defer ch.Close()

	for i := 0; i < mq.numQueue; i++ {
		if mq.topic.Queues[i] != "" {
			_, err := ch.QueueDelete(mq.topic.Queues[i], false, false, false)
			if err == nil {
				log.Printf("queue %v deleted\n", mq.topic.Queues[i])
			}
		}
	}
}
