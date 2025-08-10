package providers

import (
	"context"
	"duolingo/libraries/telemetry/otel_wrapper/trace"
	"duolingo/services/trace_service"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type MessageQueuesProvider struct {
}

func (provider *MessageQueuesProvider) Bootstrap(bootstrapCtx context.Context, scope string) {

	/* Tracing Instrumentation */

	trace.GetManager().Decorate("mq.publisher.publish(<topic>)", func(
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

	trace.GetManager().Decorate("mq.consumer.receive(<queue>)", func(
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
}

func (provider *MessageQueuesProvider) Shutdown(shutdownCtx context.Context) {
}
