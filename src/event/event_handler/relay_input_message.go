package event_handler

import (
	"sync"

	cnst "duolingo/constant"
	ed "duolingo/event/event_data"
	"duolingo/lib/event"
	"duolingo/lib/log"
	sv "duolingo/lib/service_container"
	ldt "duolingo/model/log/detail"

	"github.com/google/uuid"
)

const (
	RELAY_INP_MESG_BEGIN = "relay_input_message_begin"
	RELAY_INP_MESG_END   = "relay_input_message_end"
)

type RelayInputMessage struct {
	id        string
	container *sv.ServiceContainer
	events    *event.EventPublisher
	logger    *log.Logger
}

func NewRelayInpMsg() *RelayInputMessage {
	container := sv.GetContainer()
	return &RelayInputMessage{
		id:        uuid.New().String(),
		container: container,
		events:    container.Resolve("event.publisher").(*event.EventPublisher),
		logger:    container.Resolve("server.logger").(*log.Logger),
	}
}

func (e *RelayInputMessage) SubcriberId() string {
	return e.id
}

func (e *RelayInputMessage) Notified(wg *sync.WaitGroup, topic string, data any) {
	switch topic {
	case RELAY_INP_MESG_BEGIN:
		e.handleRelayBegin(data)
	case RELAY_INP_MESG_END:
		e.handleRelayEnd(data)
	}
}

func (e *RelayInputMessage) handleRelayBegin(data any) {
	evtData := data.(*ed.RelayInputMessage)

	var wg *sync.WaitGroup
	e.events.Notify(wg, SERVICE_OPERATION_TRACE_BEGIN, &ed.ServiceOperationTrace{
		ServiceOpt: cnst.RELAY_INP_MESG,
		OptId:      evtData.OptId,
		ParentSpan: evtData.PushNoti.Trace,
	})
	wg.Wait()

	e.events.Notify(nil, SERVICE_OPERATION_METRIC_BEGIN, &ed.ServiceOperationMetric{
		ServiceOpt: cnst.RELAY_INP_MESG,
		OptId:      evtData.OptId,
	})
}

func (e *RelayInputMessage) handleRelayEnd(data any) {
	evtData := data.(*ed.RelayInputMessage)
	traceEvtData := e.container.Resolve("events.data.sv_opt_trace." + evtData.OptId).(*ed.ServiceOperationTrace)
	metricEvtData := e.container.Resolve("events.data.sv_opt_metric." + evtData.OptId).(*ed.ServiceOperationMetric)

	var wg *sync.WaitGroup
	e.events.Notify(wg, SERVICE_OPERATION_TRACE_END, traceEvtData)
	e.events.Notify(wg, SERVICE_OPERATION_METRIC_BEGIN, metricEvtData)
	wg.Wait()

	trace := traceEvtData.Span
	if evtData.Success {
		e.logger.Info("").Detail(ldt.RelayInpMsgDetail(evtData, trace)).Write()
	} else {
		e.logger.Error("", evtData.Error).Detail(ldt.RelayInpMsgDetail(evtData, trace)).Write()
	}
	metric, err := metricEvtData.Collector.Fetch()
	if err != nil {
		e.logger.Debug("").Detail(ldt.SvOptMetricDetail(trace, metric))
	}
}
