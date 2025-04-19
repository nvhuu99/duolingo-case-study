package event_data

import (
	lc "duolingo/model/log/context"
)

type ServiceOperationTrace struct {
	OptId       string
	ServiceOpt  string
	ServiceName string
	ServiceType string
	ParentSpan  *lc.TraceSpan
	Span        *lc.TraceSpan
}
