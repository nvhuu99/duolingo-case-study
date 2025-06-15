package rabbitmq

import (
	"context"
	"testing"

	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	facade "duolingo/libraries/connection_manager/facade"
	"duolingo/libraries/pub_sub"
	"duolingo/libraries/pub_sub/drivers/rabbitmq"
	"duolingo/libraries/pub_sub/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestPubSub(t *testing.T) {
	provider := facade.Provider(context.Background()).InitRabbitMQ(connection.
		DefaultRabbitMQConnectionArgs().
		SetCredentials("root", "12345"),
	)
	publisher := rabbitmq.NewRabbitMQPublisher(provider.GetRabbitMQClient())
	subscribers := []pub_sub.Subscriber{
		rabbitmq.NewRabbitMQSubscriber(provider.GetRabbitMQClient()),
		rabbitmq.NewRabbitMQSubscriber(provider.GetRabbitMQClient()),
		rabbitmq.NewRabbitMQSubscriber(provider.GetRabbitMQClient()),
	}

	suite.Run(t, test_suites.NewPubSubTestSuite(publisher, subscribers))
}
