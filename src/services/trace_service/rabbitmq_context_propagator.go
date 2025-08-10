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

/*
Implement events.Decorator interface, allow this function to be called
just before the message_queue.Consumer calls the process message function.
Here, the context propgation is performed.
*/
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
