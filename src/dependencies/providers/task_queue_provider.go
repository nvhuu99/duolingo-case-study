package providers

import (
	"context"
	"fmt"

	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	"duolingo/libraries/message_queue/drivers/rabbitmq/task_queue"
	"duolingo/libraries/telemetry/otel_wrapper/trace"

	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type TaskQueueProvider struct {
}

func (provider *TaskQueueProvider) Bootstrap(bootstrapCtx context.Context, scope string) {

	/* Declare task queues */

	provider.declareTaskQueue(
		bootstrapCtx,
		"push_notifications",
		"push_notifications_producer",
		"push_notifications_consumer",
	)

	/* Tracing Instrumentation */

	trace.GetManager().Decorate("task_queue.producer.push(<task_queue>)", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "producer"),
			attribute.String("task_queue.driver", "rabbitmq"),
			attribute.String("task_queue.task_queue", data.Get("task_queue")),
		)
	})

	trace.GetManager().Decorate("task_queue.comsumer.comsume(<task_queue>)", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "consumer"),
			attribute.String("task_queue.driver", "rabbitmq"),
			attribute.String("task_queue.task_queue", data.Get("task_queue")),
		)
	})
}

func (provider *TaskQueueProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *TaskQueueProvider) declareTaskQueue(
	ctx context.Context,
	queueName string,
	producerName string,
	consumerName string,
) {
	connections := container.MustResolve[*facade.ConnectionProvider]()

	taskQueue := task_queue.NewTaskQueue(connections.GetRabbitMQClient())
	taskQueue.SetQueue(queueName)
	if err := taskQueue.Declare(ctx); err != nil {
		panic(fmt.Errorf("failed to declare task queue %v with error: %v", queueName, err))
	}

	container.BindSingletonAlias(producerName, func(ctx context.Context) any {
		producer := task_queue.NewTaskProducer(connections.GetRabbitMQClient())
		producer.SetQueue(queueName)
		return producer
	})

	container.BindSingletonAlias(consumerName, func(ctx context.Context) any {
		consumer := task_queue.NewTaskConsumer(connections.GetRabbitMQClient())
		consumer.SetQueue(queueName)
		return consumer
	})
}
