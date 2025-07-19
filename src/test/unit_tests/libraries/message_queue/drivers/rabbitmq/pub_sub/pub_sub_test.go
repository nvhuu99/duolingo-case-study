package rabbitmq

import (
	"context"
	"testing"

	"duolingo/dependencies"
	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	rabbitmq "duolingo/libraries/message_queue/drivers/rabbitmq/pub_sub"
	"duolingo/libraries/message_queue/pub_sub/test/test_suites"
	"duolingo/test/fixtures"

	"github.com/stretchr/testify/suite"
)

func TestPubSub(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies("common", "connections")

	provider := container.MustResolve[*facade.ConnectionProvider]()

	suite.Run(t, test_suites.NewPubSubTestSuite(
		rabbitmq.NewPublisher(provider.GetRabbitMQClient()),
		rabbitmq.NewSubscriber(provider.GetRabbitMQClient()),
		rabbitmq.NewSubscriber(provider.GetRabbitMQClient()),
	))
}
