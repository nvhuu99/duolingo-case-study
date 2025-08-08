package dependencies

import (
	"context"
	events "duolingo/libraries/events/facade"
	"duolingo/libraries/telemetry/otel_wrapper/trace"
	"duolingo/services/trace_service"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type Events struct {
}

func NewEvents() *Events {
	return &Events{}
}

func (provider *Events) Shutdown(shutdownCtx context.Context) {
}

func (provider *Events) Bootstrap(bootstrapCtx context.Context, scope string) {
	eventTracer := trace_service.NewEventTracer()
	events.InitEventManager(bootstrapCtx, 5*time.Second)
	events.AddDecorator(eventTracer)
	events.Subscribe(".*", eventTracer)

	trace.InitTraceManager(
		bootstrapCtx,
		trace.WithDefaultResource("message_input"),
		trace.WithGRPCExporter("127.0.0.1:4317", false),
	)
	
	trace.GetManager().Describe("restful.<method>(<path>)", func (
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "server"),
			attribute.String("http.request.method", data.Get("method")),
			attribute.String("url.scheme", data.Get("scheme")),
			attribute.String("url.path", data.Get("path")),
			attribute.String("url.full", data.Get("full_url")),
			attribute.String("http.response.status_code", data.Get("status_code")),
			attribute.String("user_agent.original", data.Get("user_agent")),
		)
	})

	trace.GetManager().Describe("mq.publisher.publish(<topic>)", func (
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.operation.name", "publish"),
			attribute.String("messaging.destination.kind", "topic"),
			attribute.String("messaging.destination.name", data.Get("topic")),
			attribute.String("messaging.rabbitmq.routing_key", data.Get("routing_key")),
			attribute.String("messaging.producer.name", data.Get("publisher_name")),
		)
	})
	
	trace.GetManager().Describe("pub_sub.publisher.notify(<topic>)", func (
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "producer"),
			attribute.String("pub_sub.driver", "rabbitmq"),
			attribute.String("pub_sub.topic", data.Get("topic")),
		)
	})
}
