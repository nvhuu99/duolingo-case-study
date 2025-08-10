package events

import (
	"context"
	"time"
)

/* Event */

type EventStatus string

const (
	EventStarted    EventStatus = "started"
	EventSuccess    EventStatus = "success"
	EventFailed     EventStatus = "failed"
	EventInterupted EventStatus = "interupted"
)

type Event struct {
	ctx       context.Context
	name      string
	startedAt time.Time
	endedAt   time.Time
	status    EventStatus
	err       error
	data 	  map[string]any
}

func NewEvent(name string, data map[string]any) *Event {
	if data == nil {
		data = make(map[string]any)
	}
	evt := &Event{
		name:   name,
		status: EventStarted,
		startedAt: time.Now(),
		data: data,
	}
	return evt
}

func (e *Event) Started() bool    { return e.status == EventStarted }
func (e *Event) Succeeded() bool  { return e.status == EventSuccess && e.err == nil }
func (e *Event) Failed() bool     { return e.status == EventFailed || e.err != nil }
func (e *Event) Interupted() bool { return e.status == EventInterupted }

func (e *Event) Status() string { return string(e.status) }
func (e *Event) Name() string { return e.name }
func (e *Event) Error() error { return e.err }

func (e *Event) GetData(key string) any { return e.data[key] }
func (e *Event) SetData(key string, val any) { e.data[key] = val }
func (e *Event) GetAllData() map[string]any { return e.data }
func (e *Event) MergeData(source map[string]any) { 
	for k, v := range source {
		e.SetData(k, v)
	}
}

func (e *Event) Context() context.Context { return e.ctx }
func (e *Event) StartTime() time.Time { return e.startedAt }
func (e *Event) EndTime() time.Time { return e.endedAt }

func (e *Event) HasEnded() bool {
	return e.status == EventSuccess ||
		e.status == EventFailed ||
		e.status == EventInterupted
}

/* EventBuilder */

type EventBuilder struct {
	event *Event
}

func NewEventBuilder(event *Event) *EventBuilder {
	return &EventBuilder{
		event: event,
	}
}

func (builder *EventBuilder) SetContext(ctx context.Context) {
	builder.event.ctx = ctx
}
