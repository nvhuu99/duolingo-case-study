package events

import (
	"context"
	"time"
)

type Event struct {
	ctx       context.Context
	name      string
	startedAt time.Time
	endedAt   time.Time
}
