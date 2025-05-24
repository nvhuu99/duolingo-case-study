package result

import "time"

type WorkloadMetadataResult struct {
	TraceId string `json:"trace_id" bson:"trace_id"`
	StartTime time.Time `json:"start_time" bson:"start_time"`
	ServiceInstances []struct{
		ServiceName string `json:"service_name" bson:"service_name"`
		InstanceIds []string `json:"instance_ids" bson:"instance_ids"`
	} `json:"service_instances" bson:"service_instances"`
}