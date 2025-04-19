package event_handler

import (
	cnst "duolingo/constant"
	ed "duolingo/event/event_data"
	"duolingo/lib/event"
	"duolingo/lib/log"
	sv "duolingo/lib/service_container"
	ldt "duolingo/model/log/detail"
	"sync"

	"github.com/google/uuid"
)

const (
	INP_MSG_REQUEST_BEGIN = "input_message_request_begin"
	INP_MSG_REQUEST_END   = "input_message_request_end"
)

type InputMessageRequest struct {
	id        string
	logger    *log.Logger
	events    *event.EventPublisher
	container *sv.ServiceContainer
}

func NewInputMessage() *InputMessageRequest {
	container := sv.GetContainer()
	return &InputMessageRequest{
		id:        uuid.NewString(),
		container: container,
		logger:    container.Resolve("server.logger").(*log.Logger),
		events:    container.Resolve("event.publisher").(*event.EventPublisher),
	}
}

func (e *InputMessageRequest) SubcriberId() string {
	return e.id
}

func (e *InputMessageRequest) Notified(topic string, data any) {
	switch topic {
	case INP_MSG_REQUEST_BEGIN:
		e.handleRequestBegin(data)
	case INP_MSG_REQUEST_END:
		e.handleRequestEnd(data)
	}
}

func (e *InputMessageRequest) handleRequestBegin(data any) {
	evtData := data.(*ed.InputMessageRequest)

	e.events.Notify(nil, SERVICE_OPERATION_TRACE_BEGIN, &ed.ServiceOperationTrace{
		ServiceName: cnst.SV_INP_MESG,
		ServiceType: cnst.ServiceTypes[cnst.SV_INP_MESG],
		ServiceOpt:  cnst.INP_MESG_REQUEST,
		OptId:       evtData.OptId,
		ParentSpan:  evtData.PushNoti.Trace,
	})

	e.events.Notify(nil, SERVICE_OPERATION_METRIC_BEGIN, &ed.ServiceOperationMetric{
		ServiceName: cnst.SV_INP_MESG,
		ServiceType: cnst.ServiceTypes[cnst.SV_INP_MESG],
		ServiceOpt:  cnst.INP_MESG_REQUEST,
		OptId:       evtData.OptId,
	})
}

func (e *InputMessageRequest) handleRequestEnd(data any) {
	evtData := data.(*ed.InputMessageRequest)
	traceEvtData := e.container.Resolve("events.data.sv_opt_trace." + evtData.OptId).(*ed.ServiceOperationTrace)
	metricEvtData := e.container.Resolve("events.data.sv_opt_metric." + evtData.OptId).(*ed.ServiceOperationMetric)

	wg := new(sync.WaitGroup)
	e.events.Notify(wg, SERVICE_OPERATION_TRACE_END, traceEvtData)
	e.events.Notify(wg, SERVICE_OPERATION_METRIC_END, metricEvtData)
	wg.Wait()

	trace := traceEvtData.Span
	if evtData.Success {
		e.logger.Info("").Detail(ldt.InpMsgRequestDetail(evtData, trace)).Write()
	} else {
		e.logger.Error("", evtData.Error).Detail(ldt.InpMsgRequestDetail(evtData, trace)).Write()
	}
	metric, _ := metricEvtData.Collector.Fetch()
	if metric != nil {
		e.logger.Debug("").Detail(ldt.SvOptMetricDetail(trace, metric)).Write()
	}
}
