package event_handler

import (
	"sync"

	cnst "duolingo/constant"
	ed "duolingo/event/event_data"
	"duolingo/lib/event"
	log "duolingo/lib/log"
	sv "duolingo/lib/service_container"
	ldt "duolingo/model/log/detail"

	"github.com/google/uuid"
)

const (
	BUILD_PUSH_NOTI_MESG_BEGIN = "build_push_notification_message_begin"
	BUILD_PUSH_NOTI_MESG_END   = "build_push_notification_message_end"
)

type BuildPushNotiMessage struct {
	id        string
	logger    *log.Logger
	container *sv.ServiceContainer
	events    *event.EventPublisher
}

func NewBuildPushNotiMsg() *BuildPushNotiMessage {
	container := sv.GetContainer()
	return &BuildPushNotiMessage{
		id:        uuid.NewString(),
		container: container,
		logger:    container.Resolve("server.logger").(*log.Logger),
		events:    container.Resolve("event.publisher").(*event.EventPublisher),
	}
}

func (e *BuildPushNotiMessage) SubcriberId() string {
	return e.id
}

func (e *BuildPushNotiMessage) Notified(wg *sync.WaitGroup, topic string, data any) {
	switch topic {
	case BUILD_PUSH_NOTI_MESG_BEGIN:
		e.handleBuildBegin(data)
	case BUILD_PUSH_NOTI_MESG_END:
		e.handleBuildEnd(data)
	}
}

func (e *BuildPushNotiMessage) handleBuildBegin(data any) {
	evtData := data.(*ed.BuildPushNotiMessage)

	var wg *sync.WaitGroup
	e.events.Notify(wg, SERVICE_OPERATION_TRACE_BEGIN, &ed.ServiceOperationTrace{
		ServiceOpt: cnst.BUILD_PUSH_NOTI_MESG,
		OptId:      evtData.OptId,
		ParentSpan: evtData.PushNoti.Trace,
	})
	wg.Wait()

	e.events.Notify(nil, SERVICE_OPERATION_METRIC_BEGIN, &ed.ServiceOperationMetric{
		ServiceOpt: cnst.BUILD_PUSH_NOTI_MESG,
		OptId:      evtData.OptId,
	})
}

func (e *BuildPushNotiMessage) handleBuildEnd(data any) {
	evtData := data.(*ed.BuildPushNotiMessage)
	traceEvtData := e.container.Resolve("events.data.sv_opt_trace." + evtData.OptId).(*ed.ServiceOperationTrace)
	metricEvtData := e.container.Resolve("events.data.sv_opt_metric." + evtData.OptId).(*ed.ServiceOperationMetric)

	var wg *sync.WaitGroup
	e.events.Notify(wg, SERVICE_OPERATION_TRACE_END, traceEvtData)
	e.events.Notify(wg, SERVICE_OPERATION_METRIC_BEGIN, metricEvtData)
	wg.Wait()

	trace := traceEvtData.Span
	if evtData.Success {
		e.logger.Info("").Detail(ldt.BuildNotificationDetail(evtData, trace)).Write()
	} else {
		e.logger.Error("", evtData.Error).Detail(ldt.BuildNotificationDetail(evtData, trace)).Write()
	}
	metric, err := metricEvtData.Collector.Fetch()
	if err != nil {
		e.logger.Debug("").Detail(ldt.SvOptMetricDetail(trace, metric))
	}
}
