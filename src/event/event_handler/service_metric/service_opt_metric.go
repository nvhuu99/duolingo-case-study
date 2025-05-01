package event_handler

import (
	"context"
	"time"

	ed "duolingo/event/event_data"

	config "duolingo/lib/config_reader"
	"duolingo/lib/event"
	mtr "duolingo/lib/metric"
	sv "duolingo/lib/service_container"
	collector "duolingo/event/event_handler/service_metric/stats_collector"

	"github.com/google/uuid"
)

const (
	SERVICE_OPERATION_METRIC_BEGIN = "service_operation_metric_begin"
	SERVICE_OPERATION_METRIC_END   = "service_operation_metric_end"
)

type ServiceOperationMetric struct {
	id        string
	container *sv.ServiceContainer
	ctx context.Context
	conf config.ConfigReader
	events *event.EventPublisher
}

func NewSvOptMetric() *ServiceOperationMetric {
	container := sv.GetContainer()
	conf := container.Resolve("config").(config.ConfigReader)
	ctx := container.Resolve("server.ctx").(context.Context)
	eventPublisher := container.Resolve("event.publisher").(*event.EventPublisher)
	return &ServiceOperationMetric{
		id:        uuid.New().String(),
		container: container,
		conf: conf,
		ctx: ctx,
		events: eventPublisher,
	}
}

func (e *ServiceOperationMetric) SubscriberId() string {
	return e.id
}

func (e *ServiceOperationMetric) Notified(topic string, data any) {
	switch topic {
	case SERVICE_OPERATION_METRIC_BEGIN:
		e.handleServiceOperationBegin(data)
	case SERVICE_OPERATION_METRIC_END:
		e.handleServiceOperationEnd(data)
	}
}

func (e *ServiceOperationMetric) handleServiceOperationBegin(data any) {	
	evtData := data.(*ed.ServiceOperationMetric)
	e.container.Bind("events.data.sv_opt_metric."+evtData.OptId, func() any { return evtData })
	
	rabbitmqStats := e.container.Resolve("metric.rabbitmq_stats_collector").(*collector.RabbitMQStatsCollector)
	dpInterval := e.conf.GetInt(evtData.ServiceName+".metric.datapoint_interval_ms", 15000)
	sInterval := e.conf.GetInt(evtData.ServiceName+".metric.snapshot_interval_ms", 100)
	evtData.Metric = mtr.NewMetric(e.ctx, 
		time.Duration(dpInterval)*time.Millisecond, 
		time.Duration(sInterval)*time.Millisecond,
	)
	evtData.Metric.WithCollector("system_stats", new(collector.SystemStatsCollector))
	evtData.Metric.WithCollector("message_queue_stats", rabbitmqStats)
	evtData.Metric.CaptureStart()
}

func (e *ServiceOperationMetric) handleServiceOperationEnd(data any) {
	evtData := data.(*ed.ServiceOperationMetric)
	evtData.Metric.CaptureEnd()
}
