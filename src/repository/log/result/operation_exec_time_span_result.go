package result

type OperationExecTimeSpanResult struct {
    ServiceName      string          `json:"service_name" bson:"service_name"`
    ServiceOperation string          `json:"service_operation" bson:"service_operation"`
    OperationStartLatencyMs int `json:"operation_start_latency_ms" bson:"operation_start_latency_ms"`
    OperationEndLatencyMs int `json:"operation_end_latency_ms" bson:"operation_end_latency_ms"`
}
