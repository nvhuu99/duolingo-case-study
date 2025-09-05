package providers

import (
	"context"
	"duolingo/libraries/telemetry/otel_wrapper/log"
	"duolingo/libraries/telemetry/otel_wrapper/trace"
	"duolingo/services/trace_service"

	container "duolingo/libraries/dependencies_container"
	event "duolingo/libraries/events"
	events "duolingo/libraries/events/facade"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type MessageQueuesProvider struct {
}

func (provider *MessageQueuesProvider) Bootstrap(bootstrapCtx context.Context, scope string) {

	tracer := container.MustResolve[*trace.TraceManager]()
	logger := container.MustResolve[*log.Logger]()

	/* Tracing Instrumentation */

	tracer.Decorate("mq.publisher.publish(<topic>)", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.operation.name", "publish"),
			attribute.String("messaging.destination.kind", "topic"),
			attribute.String("messaging.destination.name", data.Get("topic")),
			attribute.String("messaging.rabbitmq.routing_key", data.Get("routing_key")),
		)

		// Inject the span context propagation data into the message headers
		// before it's sent to the broker.
		propagationCtx := otlptrace.ContextWithSpan(context.Background(), span)
		if headers, ok := data.GetAny("message_headers").(amqp091.Table); ok {
			carrier := trace_service.AMQPHeadersCarrier(headers)
			otel.GetTextMapPropagator().Inject(propagationCtx, carrier)
		}
	})

	tracer.Decorate("mq.consumer.receive(<queue>)", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.operation.name", "receive"),
			attribute.String("messaging.source.kind", "queue"),
			attribute.String("messaging.source.name", data.Get("queue")),
		)
	})

	/* Logs Instrumentation */

	events.SubscribeFunc("mq.publisher.publish", func(e *event.Event) {
		logger.Write(logger.
			UnlessError(
				e.Error(), "failed to published message",
				log.LevelInfo, "message published",
			).
			Data(map[string]any{
				"messaging.system":               "rabbitmq",
				"messaging.operation.name":       "publish",
				"messaging.destination.kind":     "topic",
				"messaging.destination.name":     e.GetData("topic"),
				"messaging.rabbitmq.routing_key": e.GetData("routing_key"),
			}),
		)
	})

	events.SubscribeFunc("mq.consumer.receive", func(e *event.Event) {
		logger.Write(logger.
			UnlessError(
				e.Error(), "failed to process message",
				log.LevelInfo, "message proccessed",
			).
			Data(map[string]any{
				"messaging.system":           "rabbitmq",
				"messaging.operation.name":   "receive",
				"messaging.destination.kind": "queue",
				"messaging.destination.name": e.GetData("queue"),
			}),
		)
	})
}

func (provider *MessageQueuesProvider) Shutdown(shutdownCtx context.Context) {
}
