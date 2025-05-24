package result

type WorkloadMetricSummaryResult struct {
	MetricTarget string `json:"metric_target" bson:"metric_target"`
	MetricName string `json:"metric_name" bson:"metric_name"`
	Summary struct{
		Average float64 `json:"average" bson:"average"`
		Median float64 `json:"median" bson:"median"`
		Minimum float64 `json:"minimum" bson:"minimum"`
		Maximum float64 `json:"maximum" bson:"maximum"`
		P5 float64 `json:"p5" bson:"p5"`
		P25 float64 `json:"p25" bson:"p25"`
		P75 float64 `json:"p75" bson:"p75"`
		P95 float64 `json:"p95" bson:"p95"`
	} `json:"summary" bson:"summary"`
}