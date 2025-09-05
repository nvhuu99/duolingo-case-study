package trace_service

import (
	container "duolingo/libraries/dependencies_container"
	"duolingo/libraries/events"
	trace "duolingo/libraries/telemetry/otel_wrapper/trace"

	"go.opentelemetry.io/otel/codes"
)

type GlobalTracer struct {
	*events.BaseEventSubscriber
	*trace.TraceManager
}

func NewGlobalTracer() *GlobalTracer {
	return &GlobalTracer{
		BaseEventSubscriber: events.NewBaseEventSubscriber(),
		TraceManager:        container.MustResolve[*trace.TraceManager](),
	}
}

/*
Implement events.Decorator interface, allow the GlobalTracer to create
a trace-span immediately when an event has started.
*/
func (tracer *GlobalTracer) Decorate(event *events.Event, builder *events.EventBuilder) {
	spanCtx, _ := tracer.Start(
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
Implement events.Subscriber interface, allow the GlobalTracer
to end a trace-span exactly when an an events (and it's childs) has ended
and collected by the events.EventManager.
*/
func (tracer *GlobalTracer) Notify(event *events.Event) {
	// EventManager notify on event start, skip
	if !event.HasEnded() {
		return
	}

	span := tracer.Span(event.Context())

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

	tracer.End(
		span,
		event.EndTime(),
		statusCode,
		message,
		event.Error(),
		trace.NewDataBag().Merge(event.GetAllData()),
	)
}
