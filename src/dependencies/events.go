package dependencies

import (
	"context"
	events "duolingo/libraries/events/facade"
	"duolingo/libraries/telemetry/otel_wrapper/trace"
	"duolingo/services/trace_service"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
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
	rabbitMQPropagator := trace_service.NewRabbitMQContextPropagator()
	
	events.InitEventManager(bootstrapCtx, 5*time.Second)
	events.AddDecorators(
		rabbitMQPropagator,
		eventTracer,
	)
	events.Subscribe(".*", eventTracer)

	trace.InitTraceManager(
		bootstrapCtx,
		trace.WithDefaultResource("message_input"),
		trace.WithGRPCExporter("127.0.0.1:4317", false),
	)
	
	trace.GetManager().Decorate("restful.<method>(<path>)", func (
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

	provider.describeMessageQueueTraces()
}

func (provider *Events) describeMessageQueueTraces() {
	trace.GetManager().Decorate("mq.publisher.publish(<topic>)", func (
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
		// Context propagation
		spanPropagationCtx := otlptrace.ContextWithSpan(
			context.Background(), 
			otlptrace.SpanFromContext(data.GetAny("parent_ctx").(context.Context)),
		)
		if headers, ok := data.GetAny("message_headers").(amqp091.Table); ok {
			carrier := trace_service.AMQPHeadersCarrier(headers)
			otel.GetTextMapPropagator().Inject(spanPropagationCtx, carrier)
		}
	})

	trace.GetManager().Decorate("mq.consumer.receive(<queue>)", func (
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

	trace.GetManager().Decorate("mq.consumer.ack(<queue>, <action>)", func (
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.operation.name", "ack"),
			attribute.String("messaging.source.kind", "queue"),
			attribute.String("messaging.source.name", data.Get("queue")),
			attribute.String("mq.consumer.comsume_action", data.Get("action")),
		)
	})
	
	trace.GetManager().Decorate("pub_sub.publisher.notify(<topic>)", func (
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "producer"),
			attribute.String("pub_sub.driver", "rabbitmq"),
			attribute.String("pub_sub.topic", data.Get("topic")),
		)
	})

	trace.GetManager().Decorate("task_queue.producer.push(<task_queue>)", func (
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "producer"),
			attribute.String("task_queue.driver", "rabbitmq"),
			attribute.String("task_queue.task_queue", data.Get("task_queue")),
		)
	})

	trace.GetManager().Decorate("task_queue.comsumer.comsume(<task_queue>)", func (
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

