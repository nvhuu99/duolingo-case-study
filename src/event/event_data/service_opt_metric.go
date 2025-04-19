package event_data

import "duolingo/lib/metric"

type ServiceOperationMetric struct {
	OptId       string
	ServiceOpt  string
	ServiceName string
	ServiceType string
	Collector   *metric.MetricCollector
}
