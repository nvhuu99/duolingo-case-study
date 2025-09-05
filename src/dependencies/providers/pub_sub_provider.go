package providers

import (
	"context"
	"fmt"

	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	event "duolingo/libraries/events"
	events "duolingo/libraries/events/facade"
	"duolingo/libraries/message_queue/drivers/rabbitmq/pub_sub"
	"duolingo/libraries/telemetry/otel_wrapper/log"
	"duolingo/libraries/telemetry/otel_wrapper/trace"

	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type PubSubProvider struct {
}

func (provider *PubSubProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	tracer := container.MustResolve[*trace.TraceManager]()
	logger := container.MustResolve[*log.Logger]()

	/* Declare publishers and subscribers */

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

	/* Tracing Instrumentation */

	tracer.Decorate("pub_sub.publisher.notify(<topic>)", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "producer"),
			attribute.String("pub_sub.driver", "rabbitmq"),
			attribute.String("pub_sub.topic", data.Get("topic")),
		)
	})

	tracer.Decorate("pub_sub.subscriber.notified(<topic>)", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "consumer"),
			attribute.String("pub_sub.driver", "rabbitmq"),
			attribute.String("pub_sub.topic", data.Get("topic")),
		)
	})

	/* Logs Instrumentation */

	events.SubscribeFunc("pub_sub.publisher.notify", func(e *event.Event) {
		logger.Write(logger.
			UnlessError(
				e.Error(), "failed to notify message",
				log.LevelInfo, "message notified",
			).
			Data(map[string]any{
				"pub_sub.driver": "rabbitmq",
				"pub_sub.topic":  e.GetData("topic"),
			}),
		)
	})

	events.SubscribeFunc("pub_sub.subscriber.notified", func(e *event.Event) {
		logger.Write(logger.
			UnlessError(
				e.Error(), "failed to notify message",
				log.LevelInfo, "message notified",
			).
			Data(map[string]any{
				"pub_sub.driver": "rabbitmq",
				"pub_sub.topic":  e.GetData("topic"),
			}),
		)
	})
}

func (provider *PubSubProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *PubSubProvider) declareTopic(
	topicName string,
	publisherName string,
	subscriberName string,
) {
	connections := container.MustResolve[*facade.ConnectionProvider]()

	container.BindSingletonAlias(publisherName, func(ctx context.Context) any {
		publisher := pub_sub.NewPublisher(connections.GetRabbitMQClient())
		publisher.SetMainTopic(topicName)
		if declareErr := publisher.DeclareMainTopic(ctx); declareErr != nil {
			panic(fmt.Errorf("failed to declare topic %v with error: %v", topicName, declareErr))
		}
		return publisher
	})

	container.BindSingletonAlias(subscriberName, func(ctx context.Context) any {
		subscriber := pub_sub.NewSubscriber(connections.GetRabbitMQClient())
		subscriber.SetMainTopic(topicName)
		if subscribeErr := subscriber.SubscribeMainTopic(ctx); subscribeErr != nil {
			panic(fmt.Errorf("failed to subscribe topic %v with error: %v", topicName, subscribeErr))
		}
		return subscriber
	})
}
