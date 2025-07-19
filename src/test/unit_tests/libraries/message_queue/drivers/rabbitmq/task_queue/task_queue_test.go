package task_queue

import (
	"context"
	"testing"

	"duolingo/dependencies"
	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq/task_queue"
	"duolingo/libraries/message_queue/task_queue/test/test_suites"
	"duolingo/test/fixtures"

	"github.com/stretchr/testify/suite"
)

func TestPubSub(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies("test", []string{
		"common",
		"connections",
	})

	provider := container.MustResolve[*facade.ConnectionProvider]()

	suite.Run(t, test_suites.NewTaskQueueTestSuite(
		driver.NewTaskQueue(provider.GetRabbitMQClient()),
		driver.NewTaskProducer(provider.GetRabbitMQClient()),
		driver.NewTaskConsumer(provider.GetRabbitMQClient()),
		driver.NewTaskConsumer(provider.GetRabbitMQClient()),
	))
}
