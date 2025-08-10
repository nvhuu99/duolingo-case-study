package trace_service

import (
	"duolingo/libraries/events"
	"strings"

	"github.com/rabbitmq/amqp091-go"

	"go.opentelemetry.io/otel"
)

type RabbitMQContextPropagator struct {
}

func NewRabbitMQContextPropagator() *RabbitMQContextPropagator {
	return &RabbitMQContextPropagator{}
}

func (propagator *RabbitMQContextPropagator) Decorate(
	event *events.Event,
	builder *events.EventBuilder,
) {
	if strings.HasPrefix(event.Name(), "mq.consumer.receive") {
		if headers, ok := event.GetData("message_headers").(amqp091.Table); ok {
			carrier := AMQPHeadersCarrier(headers)
			propagatexCtx := otel.GetTextMapPropagator().Extract(event.Context(), carrier)
			builder.SetContext(propagatexCtx)
		}
	}
}
