package task_queue

import (
	"context"
	"testing"

	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	facade "duolingo/libraries/connection_manager/facade"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq/task_queue"
	"duolingo/libraries/message_queue/task_queue/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestPubSub(t *testing.T) {
	provider := facade.Provider(context.Background()).InitRabbitMQ(connection.
		DefaultRabbitMQConnectionArgs().
		SetCredentials("root", "12345"),
	)

	suite.Run(t, test_suites.NewTaskQueueTestSuite(
		driver.NewTaskQueue(provider.GetRabbitMQClient()),
		driver.NewTaskProducer(provider.GetRabbitMQClient()),
		driver.NewTaskConsumer(provider.GetRabbitMQClient()),
		driver.NewTaskConsumer(provider.GetRabbitMQClient()),
	))
}
