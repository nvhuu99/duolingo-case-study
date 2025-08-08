package trace_service

import (
	"duolingo/libraries/events"
	trace "duolingo/libraries/telemetry/otel_wrapper/trace"

	"go.opentelemetry.io/otel/codes"
)

type EventTracer struct {
	*events.BaseEventSubscriber
}

func NewEventTracer() *EventTracer {
	return &EventTracer{
		BaseEventSubscriber: events.NewBaseEventSubscriber(),
	}
}

func (tracer *EventTracer) Decorate(event *events.Event, builder *events.EventBuilder) {
	spanCtx, _ := trace.GetManager().Start(
		event.GetContext(), 
		event.Name(),
		event.StartTime(),
	)

	builder.SetContext(spanCtx)
}

func (tracer *EventTracer) Notify(event *events.Event) {
	if ! event.HasEnded() {
		return
	}

	span := trace.GetManager().Span(event.GetContext())

	var statusCode codes.Code
	var message string
	
	if event.Interupted() {
		statusCode = codes.Unset
		message = "span execution has been interupted"
	} else if event.Failed() {
		statusCode = codes.Error
	} else {
		statusCode = codes.Ok
	}

	trace.GetManager().End(
		span,
		event.EndTime(),
		statusCode,
		message,
		event.Error(),
		trace.NewDataBag().Merge(event.GetAllData()),
	)
}
