package result

type WorkloadOperationListResult struct {
	ServiceName string `json:"service_name" bson:"service_name"`
	ServiceOperation string `json:"service_operation" bson:"service_operation"`
}
