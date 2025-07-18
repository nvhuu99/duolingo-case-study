package rabbitmq

import (
	"context"
	"testing"

	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	facade "duolingo/libraries/connection_manager/facade"
	rabbitmq "duolingo/libraries/message_queue/drivers/rabbitmq/pub_sub"
	"duolingo/libraries/message_queue/pub_sub/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestPubSub(t *testing.T) {
	provider := facade.Provider(context.Background()).InitRabbitMQ(connection.
		DefaultRabbitMQConnectionArgs().
		SetCredentials("root", "12345"),
	)

	suite.Run(t, test_suites.NewPubSubTestSuite(
		rabbitmq.NewPublisher(provider.GetRabbitMQClient()),
		rabbitmq.NewSubscriber(provider.GetRabbitMQClient()),
		rabbitmq.NewSubscriber(provider.GetRabbitMQClient()),
	))
}
