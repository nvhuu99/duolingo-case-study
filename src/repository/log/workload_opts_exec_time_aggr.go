package log

type WorkloadOptsExecTimeAggr struct {
    ServiceName      string          `json:"service_name" bson:"service_name"`
    ServiceOperation string          `json:"service_operation" bson:"service_operation"`
    StartTimeLatency struct {
		Min int `json:"min" bson:"min"`
	} `json:"start_time_latency" bson:"start_time_latency"`
    Duration         struct {
		Count      int     `json:"count" bson:"count"`
		Min        int     `json:"min" bson:"min"`
		Max        int     `json:"max" bson:"max"`
		Avg        float64 `json:"avg" bson:"avg"`
		Median     int     `json:"median" bson:"median"`
		Percentile5 int    `json:"percentile_5" bson:"percentile_5"`
		Percentile25 int   `json:"percentile_25" bson:"percentile_25"`
		Percentile50 int   `json:"percentile_50" bson:"percentile_50"`
		Percentile75 int   `json:"percentile_75" bson:"percentile_75"`
		Percentile95 int   `json:"percentile_95" bson:"percentile_95"`
	} `json:"duration" bson:"duration"`
}
