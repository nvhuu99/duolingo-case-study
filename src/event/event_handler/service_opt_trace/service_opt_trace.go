package event_handler

import (
	"time"

	ed "duolingo/event/event_data"
	sv "duolingo/lib/service_container"
	lc "duolingo/model/log/context"

	"github.com/google/uuid"
)

const (
	SERVICE_OPERATION_TRACE_BEGIN = "service_operation_trace_begin"
	SERVICE_OPERATION_TRACE_END   = "service_operation_trace_end"
)

type ServiceOperationTrace struct {
	id        string
	container *sv.ServiceContainer
}

func NewSvOptTrace() *ServiceOperationTrace {
	return &ServiceOperationTrace{
		id:        uuid.New().String(),
		container: sv.GetContainer(),
	}
}

func (e *ServiceOperationTrace) SubscriberId() string {
	return e.id
}

func (e *ServiceOperationTrace) Notified(topic string, data any) {
	switch topic {
	case SERVICE_OPERATION_TRACE_BEGIN:
		e.handleServiceOperationBegin(data)
	case SERVICE_OPERATION_TRACE_END:
		e.handleServiceOperationEnd(data)
	}
}

func (e *ServiceOperationTrace) handleServiceOperationBegin(data any) {
	evtData := data.(*ed.ServiceOperationTrace)
	evtData.Span = &lc.TraceSpan{
		TraceId:          evtData.ParentSpan.TraceId,
		ParentSpanId:     evtData.ParentSpan.SpanId,
		SpanId:           uuid.NewString(),
		ServiceName:      evtData.ServiceName,
		ServiceOperation: evtData.ServiceOpt,
		ServiceType:      evtData.ServiceType,
		StartTime:        time.Now(),
	}
	e.container.BindSingleton("events.data.sv_opt_trace."+evtData.OptId, func() any {
		return evtData
	})
}

func (e *ServiceOperationTrace) handleServiceOperationEnd(data any) {
	evtData := data.(*ed.ServiceOperationTrace)
	evtData.Span.EndTime = time.Now()
}
