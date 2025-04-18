package context

import "time"

type TraceSpan struct {
	TraceId      string `json:"trace_id"`
	ParentSpanId string `json:"parent_span_id"`
	SpanId       string `json:"span_id"`

	ServiceName      string `json:"service_name"`
	ServiceOperation string `json:"service_operation"`
	ServiceType      string `json:"service_type"`
	InstanceId       string `json:"instance_id"`
	InstanceAddress  string `json:"instance_address"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
