package rabbitmq

import (
	"context"
	"duolingo/libraries/connection_manager"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	*connection_manager.Client

	declareTimeout time.Duration
}

func (client *RabbitMQClient) ExecuteClosure(
	timeout time.Duration,
	closure func(ctx context.Context, ch *amqp.Channel) error,
) error {
	wrapper := func(ctx context.Context, conn any) error {
		channel, _ := conn.(*amqp.Channel)
		return closure(ctx, channel)
	}
	return client.Client.ExecuteClosure(timeout, wrapper)
}

func (client *RabbitMQClient) GetConnection() *amqp.Channel {
	conn := client.Client.GetConnection()
	if channel, ok := conn.(*amqp.Channel); ok {
		return channel
	}
	return nil
}

func (client *RabbitMQClient) GetDeclareTimeout() time.Duration {
	return client.declareTimeout
}
