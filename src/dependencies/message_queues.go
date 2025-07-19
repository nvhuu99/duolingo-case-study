package dependencies

import (
	"context"
	"fmt"

	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	"duolingo/libraries/message_queue/drivers/rabbitmq/pub_sub"
	"duolingo/libraries/message_queue/drivers/rabbitmq/task_queue"
)

type MessageQueues struct {
	connections *facade.ConnectionProvider
}

func NewMessageQueues() *MessageQueues {
	return &MessageQueues{}
}

func (provider *MessageQueues) Shutdown() {
}

func (provider *MessageQueues) Bootstrap() {
	provider.connections = container.MustResolve[*facade.ConnectionProvider]()

	provider.declareTopic(
		"message_inputs",
		"message_input_publisher",
		"message_input_subscriber",
	)

	provider.declareTopic(
		"noti_builder_jobs",
		"noti_builder_jobs_publisher",
		"noti_builder_jobs_subscriber",
	)

	provider.declareTaskQueue(
		"push_notifications",
		"push_notifications_producer",
		"push_notifications_consumer",
	)
}

func (provider *MessageQueues) declareTopic(
	topicName string,
	publisherName string,
	subscriberName string,
) {
	container.BindSingletonAlias(publisherName, func(ctx context.Context) any {
		publisher := pub_sub.NewPublisher(provider.connections.GetRabbitMQClient())
		publisher.SetMainTopic(topicName)
		if declareErr := publisher.DeclareMainTopic(); declareErr != nil {
			panic(fmt.Errorf("failed to declare topic %v with error: %v", topicName, declareErr))
		}
		return publisher
	})

	container.BindSingletonAlias(subscriberName, func(ctx context.Context) any {
		subscriber := pub_sub.NewSubscriber(provider.connections.GetRabbitMQClient())
		subscriber.SetMainTopic(topicName)
		if subscribeErr := subscriber.SubscribeMainTopic(); subscribeErr != nil {
			panic(fmt.Errorf("failed to subscribe topic %v with error: %v", topicName, subscribeErr))
		}
		return subscriber
	})
}

func (provider *MessageQueues) declareTaskQueue(
	queueName string,
	producerName string,
	consumerName string,
) {
	taskQueue := task_queue.NewTaskQueue(provider.connections.GetRabbitMQClient())
	if err := taskQueue.Declare(); err != nil {
		panic(fmt.Errorf("failed to declare task queue %v with error: %v", queueName, err))
	}

	container.BindSingletonAlias(producerName, func(ctx context.Context) any {
		producer := task_queue.NewTaskProducer(provider.connections.GetRabbitMQClient())
		producer.SetQueue(queueName)
		return producer
	})

	container.BindSingletonAlias(consumerName, func(ctx context.Context) any {
		consumer := task_queue.NewTaskConsumer(provider.connections.GetRabbitMQClient())
		consumer.SetQueue(queueName)
		return consumer
	})
}
