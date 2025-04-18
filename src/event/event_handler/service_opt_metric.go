package event_handler

import (
	"context"
	"strings"
	"sync"
	"time"

	ed "duolingo/event/event_data"
	config "duolingo/lib/config_reader"
	"duolingo/lib/metric"
	sv "duolingo/lib/service_container"

	"github.com/google/uuid"
)

const (
	SERVICE_OPERATION_METRIC_BEGIN = "service_operation_metric_begin"
	SERVICE_OPERATION_METRIC_END   = "service_operation_metric_end"
)

type ServiceOperationMetric struct {
	id        string
	container *sv.ServiceContainer
}

func NewSvOptMetric() *ServiceOperationMetric {
	return &ServiceOperationMetric{
		id:        uuid.New().String(),
		container: sv.GetContainer(),
	}
}

func (e *ServiceOperationMetric) SubcriberId() string {
	return e.id
}

func (e *ServiceOperationMetric) Notified(wg *sync.WaitGroup, topic string, data any) {
	switch topic {
	case SERVICE_OPERATION_METRIC_BEGIN:
		e.handleServiceOperationBegin(data)
	case SERVICE_OPERATION_METRIC_END:
		e.handleServiceOperationEnd(data)
	}
}

func (e *ServiceOperationMetric) handleServiceOperationBegin(data any) {
	evtData := data.(*ed.ServiceOperationMetric)
	svName := strings.Split(evtData.ServiceOpt, ":")[1]
	conf := e.container.Resolve("config").(config.ConfigReader)
	ctx := e.container.Resolve("server.ctx").(context.Context)
	dpInterval := conf.GetInt(svName+".metric.datapoint_interval_ms", 15000)
	sInterval := conf.GetInt(svName+".metric.snapshot_interval_ms", 100)
	evtData.Collector = metric.NewCollector(ctx,
		time.Duration(dpInterval)*time.Millisecond,
		time.Duration(sInterval)*time.Millisecond,
	)
	e.container.Bind("events.data.sv_opt_metric."+evtData.OptId, func() any {
		return evtData
	})
	evtData.Collector.CaptureStart(metric.CaptureAll)
}

func (e *ServiceOperationMetric) handleServiceOperationEnd(data any) {
	evtData := data.(*ed.ServiceOperationMetric)
	evtData.Collector.CaptureEnd()
}
