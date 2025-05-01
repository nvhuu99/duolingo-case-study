package detail

import (
	"duolingo/lib/log"
	"duolingo/lib/metric"
	lc "duolingo/model/log/context"
)

type ServiceOperationMetric struct {
	log.Log

	LogContext struct {
		Trace *lc.TraceSpan `json:"trace"`
	} `json:"context"`

	LogData struct {
		Metric *metric.DataPoint `json:"metric"`
	} `json:"data"`
}

func SvOptMetricDetail(trace *lc.TraceSpan, metric *metric.DataPoint) map[string]any {
	return map[string]any{
		"context": map[string]any{
			"trace": trace,
		},
		"data": map[string]any{
			"metric": metric,
		},
	}
}
