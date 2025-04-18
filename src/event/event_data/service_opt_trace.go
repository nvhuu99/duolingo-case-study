package event_data

import (
	lc "duolingo/model/log/context"
)

type ServiceOperationTrace struct {
	ServiceOpt string
	OptId      string
	ParentSpan *lc.TraceSpan
	Span       *lc.TraceSpan
}
