package rabbitmq

import (
	ctxt "context"
	"log"

	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Topology struct {
	*connection.RabbitMQClient
}

func NewTopology(client *connection.RabbitMQClient) *Topology {
	return &Topology{
		RabbitMQClient: client,
	}
}

func (client *Topology) DeclareExchange(ctx ctxt.Context, opts *ExchangeOptions) error {
	defer log.Printf("Topology: exchange %v declared\n", opts.name)
	return client.ExecuteClosure(ctx, client.GetDeclareTimeout(), func(
		ctx ctxt.Context,
		ch *amqp.Channel,
	) error {
		return ch.ExchangeDeclare(
			opts.name,
			string(opts.kind),
			opts.durable,
			opts.autoDelete,
			false, // internal
			false, // no-wait
			nil,   // arguments
		)
	})
}

func (client *Topology) DeclareQueue(
	ctx ctxt.Context,
	queueOpts *QueueOptions,
	queueBindings *QueueBindings,
) error {
	return client.ExecuteClosure(ctx, client.GetDeclareTimeout(), func(
		timeoutCtx ctxt.Context,
		ch *amqp.Channel,
	) error {
		_, err := ch.QueueDeclare(
			queueOpts.name,
			queueOpts.durable,
			queueOpts.autoDelete,
			queueOpts.exclusive,
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return err
		}
		for _, binding := range queueBindings.bindings {
			bindErr := ch.QueueBind(
				queueOpts.name,
				binding.routingKey,
				binding.exchange,
				false, // no-wait
				nil,   // arguments
			)
			if bindErr != nil {
				ch.QueueDelete(
					queueOpts.name,
					true,  // do not delete if being used
					true,  // do not delete if not empty
					false, // no-wait
				)
				return bindErr
			}
			log.Printf(
				"Topology: queue %v binded to %x with key %v\n",
				queueOpts.name,
				binding.exchange,
				binding.routingKey,
			)
		}

		log.Printf("Topology: queue %v declared\n", queueOpts.name)

		return nil
	})
}

func (client *Topology) DeleteExchange(ctx ctxt.Context, name string) error {
	defer log.Printf("Topology: exchange %v name deleted\n", name)
	return client.ExecuteClosure(ctx, client.GetWriteTimeout(), func(
		timeoutCtx ctxt.Context,
		ch *amqp.Channel,
	) error {
		return ch.ExchangeDelete(name, false, false)
	})
}

func (client *Topology) DeleteQueue(ctx ctxt.Context, name string) error {
	defer log.Printf("Topology: queue %v name deleted\n", name)
	return client.ExecuteClosure(ctx, client.GetWriteTimeout(), func(
		timeoutCtx ctxt.Context,
		ch *amqp.Channel,
	) error {
		_, err := ch.QueueDelete(name, false, false, false)
		return err
	})
}
