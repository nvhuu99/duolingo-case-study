package trace_service

import (
	"duolingo/libraries/events"
	trace "duolingo/libraries/telemetry/otel_wrapper/trace"

	"go.opentelemetry.io/otel/codes"
)

type GlobalEventTracer struct {
	*events.BaseEventSubscriber
}

func NewGlobalEventTracer() *GlobalEventTracer {
	return &GlobalEventTracer{
		BaseEventSubscriber: events.NewBaseEventSubscriber(),
	}
}

/*
Implement events.Decorator interface, allow the GlobalEventTracer to create
a trace-span immediately when an event has started.
*/
func (tracer *GlobalEventTracer) Decorate(event *events.Event, builder *events.EventBuilder) {
	spanCtx, _ := trace.GetManager().Start(
		event.Context(),
		event.Name(),
		event.StartTime(),
		trace.NewDataBag().Merge(event.GetAllData()),
	)

	// Update the event context, so that when a child event is created,
	// it's context contains the span data, thus preserves the spans hierachy.
	builder.SetContext(spanCtx)
}

/*
Implement events.Subscriber interface, allow the GlobalEventTracer
to end a trace-span exactly when an an events (and it's childs) has ended
and collected by the events.EventManager.
*/
func (tracer *GlobalEventTracer) Notify(event *events.Event) {
	// EventManager notify on event start, skip
	if !event.HasEnded() {
		return
	}

	span := trace.GetManager().Span(event.Context())

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
