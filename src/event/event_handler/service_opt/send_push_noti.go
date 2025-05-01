package event_handler

import (
	"fmt"

	cnst "duolingo/constant"
	ed "duolingo/event/event_data"
	sm "duolingo/event/event_handler/service_metric"
	st "duolingo/event/event_handler/service_opt_trace"
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

func (e *SendPushNotification) SubscriberId() string {
	return e.id
}

func (e *SendPushNotification) Notified(topic string, data any) {
	switch topic {
	case SEND_PUSH_NOTI_BEGIN:
		e.handleSendPushNotiBegin(data)
	case SEND_PUSH_NOTI_END:
		e.handleSendPushNotiEnd(data)
	}
}

func (e *SendPushNotification) handleSendPushNotiBegin(data any) {
	evtData := data.(*ed.SendPushNotification)
	e.events.Notify(st.SERVICE_OPERATION_TRACE_BEGIN, &ed.ServiceOperationTrace{
		ServiceName: cnst.SV_PUSH_SENDER,
		ServiceType: cnst.ServiceTypes[cnst.SV_PUSH_SENDER],
		ServiceOpt:  cnst.SEND_PUSH_NOTI,
		OptId:       evtData.OptId,
		ParentSpan:  evtData.PushNoti.Trace,
	})
	e.events.Notify(sm.SERVICE_OPERATION_METRIC_BEGIN, &ed.ServiceOperationMetric{
		ServiceName: cnst.SV_PUSH_SENDER,
		ServiceType: cnst.ServiceTypes[cnst.SV_PUSH_SENDER],
		ServiceOpt:  cnst.SEND_PUSH_NOTI,
		OptId:       evtData.OptId,
	})
}

func (e *SendPushNotification) handleSendPushNotiEnd(data any) {
	evtData := data.(*ed.SendPushNotification)
	traceEvtData := e.container.Resolve("events.data.sv_opt_trace." + evtData.OptId).(*ed.ServiceOperationTrace)
	metricEvtData := e.container.Resolve("events.data.sv_opt_metric." + evtData.OptId).(*ed.ServiceOperationMetric)

	e.events.Notify(st.SERVICE_OPERATION_TRACE_END, traceEvtData)
	e.events.Notify(sm.SERVICE_OPERATION_METRIC_END, metricEvtData)

	trace := traceEvtData.Span
	if evtData.Result.Success {
		e.logger.Info("").Detail(ldt.SendPushNotiDetail(evtData, trace)).Write()
	} else {
		e.logger.Error("", evtData.Result.Error).Detail(ldt.SendPushNotiDetail(evtData, trace)).Write()
	}
	metric, _ := metricEvtData.Metric.Fetch()
	if metric != nil {
		e.logger.Debug("").Detail(ldt.SvOptMetricDetail(trace, metric)).Write()
	}

	fmt.Printf("push_noti_sent - has_err: %v - id: %v - title: %v - trace: %v\n",
		evtData.Result.Error,
		evtData.PushNoti.InputMessage.MessageId,
		evtData.PushNoti.InputMessage.Title,
		trace.TraceId,
	)
}
