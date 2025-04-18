package event_data

import "duolingo/lib/metric"

type ServiceOperationMetric struct {
	ServiceOpt string
	OptId      string
	Collector  *metric.MetricCollector
}
