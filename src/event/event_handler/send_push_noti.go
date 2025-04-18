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
	SEND_PUSH_NOTI_BEGIN = "send_push_notification_begin"
	SEND_PUSH_NOTI_END   = "send_push_notification_end"
)

type SendPushNotification struct {
	id        string
	container *sv.ServiceContainer
	events    *event.EventPublisher
	logger    *log.Logger
}

func NewSendPushNoti() *SendPushNotification {
	container := sv.GetContainer()
	return &SendPushNotification{
		id:        uuid.New().String(),
		container: container,
		events:    container.Resolve("event.publisher").(*event.EventPublisher),
		logger:    container.Resolve("server.logger").(*log.Logger),
	}
}

func (e *SendPushNotification) SubcriberId() string {
	return e.id
}

func (e *SendPushNotification) Notified(wg *sync.WaitGroup, topic string, data any) {
	switch topic {
	case SEND_PUSH_NOTI_BEGIN:
		e.handleSendPushNotiBegin(data)
	case SEND_PUSH_NOTI_END:
		e.handleSendPushNotiEnd(data)
	}
}

func (e *SendPushNotification) handleSendPushNotiBegin(data any) {
	evtData := data.(*ed.SendPushNotification)

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

func (e *SendPushNotification) handleSendPushNotiEnd(data any) {
	evtData := data.(*ed.SendPushNotification)
	traceEvtData := e.container.Resolve("events.data.sv_opt_trace." + evtData.OptId).(*ed.ServiceOperationTrace)
	metricEvtData := e.container.Resolve("events.data.sv_opt_metric." + evtData.OptId).(*ed.ServiceOperationMetric)

	var wg *sync.WaitGroup
	e.events.Notify(wg, SERVICE_OPERATION_TRACE_END, traceEvtData)
	e.events.Notify(wg, SERVICE_OPERATION_METRIC_BEGIN, metricEvtData)
	wg.Wait()

	trace := traceEvtData.Span
	if evtData.Result.Success {
		e.logger.Info("").Detail(ldt.SendPushNotiDetail(evtData, trace)).Write()
	} else {
		e.logger.Error("", evtData.Result.Error).Detail(ldt.SendPushNotiDetail(evtData, trace)).Write()
	}
	metric, err := metricEvtData.Collector.Fetch()
	if err != nil {
		e.logger.Debug("").Detail(ldt.SvOptMetricDetail(trace, metric))
	}
}
