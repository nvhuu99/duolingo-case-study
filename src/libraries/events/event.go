package events

import (
	"context"
	"time"
)
type Event struct {
	name      string

	ctx       context.Context

	startedAt time.Time
	endedAt   time.Time
}

func NewEvent(name string) *Event {
	return &Event{
		name: name,
	}
}
