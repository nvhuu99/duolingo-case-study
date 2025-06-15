package rabbitmq

import (
	"context"

	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQTopology struct {
	*connection.RabbitMQClient
}

func NewRabbitMQTopology(client *connection.RabbitMQClient) *RabbitMQTopology {
	return &RabbitMQTopology{
		RabbitMQClient: client,
	}
}

func (client *RabbitMQTopology) DeclareExchange(name string) error {
	declareTimeout := client.GetDeclareTimeout()
	return client.ExecuteClosure(declareTimeout, func(
		ctx context.Context,
		ch *amqp.Channel,
	) error {
		return ch.ExchangeDeclare(
			name,
			"fanout",
			// TODO: allow options
			false, // durable
			false, // auto-delete
			false, // internal
			false, // no-wait
			nil,   // arguments
		)
	})
}

func (client *RabbitMQTopology) DeclareQueue(
	queueName string,
	routingKey string,
	exchangeName string,
) error {
	declareTimeout := client.GetDeclareTimeout()
	return client.ExecuteClosure(declareTimeout, func(
		ctx context.Context,
		ch *amqp.Channel,
	) error {
		_, err := ch.QueueDeclare(
			queueName,
			// TODO: allow options
			false, // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return err
		}
		return ch.QueueBind(
			queueName,
			routingKey,
			exchangeName,
			false,
			nil,
		)
	})
}

func (client *RabbitMQTopology) DeleteQueue(name string) error {
	timeout := client.GetWriteTimeout()
	return client.ExecuteClosure(timeout, func(ctx context.Context, ch *amqp.Channel) error {
		_, err := ch.QueueDelete(name, false, false, false)
		return err
	})
}
