package param

import "duolingo/constant"

type WorkloadMetricQueryMongoWrapper struct {
	query *WorkloadMetricQueryParam
}

func NewWorkloadMongoMetricQuery(traceId string) *WorkloadMetricQueryMongoWrapper {
	builder := new(WorkloadMetricQueryMongoWrapper)
	builder.query = NewWorkloadMetricQueryParam(traceId).SetMetricTarget(constant.METRIC_TARGET_MONGO)
	return builder
}

func (builder *WorkloadMetricQueryMongoWrapper) AddMetricNames(metricNames... string) *WorkloadMetricQueryMongoWrapper {
	builder.query.AddMetricNames(metricNames...)
	return builder
}

func (builder *WorkloadMetricQueryMongoWrapper) SetServiceName(name string) *WorkloadMetricQueryMongoWrapper {
	builder.query.SetServiceName(name)
	return builder
}

func (builder *WorkloadMetricQueryMongoWrapper) SetServiceOperation(opt string) *WorkloadMetricQueryMongoWrapper {
	builder.query.SetServiceOperation(opt)
	return builder
}

func (builder *WorkloadMetricQueryMongoWrapper) SetQuery(query *WorkloadMetricQueryParam) *WorkloadMetricQueryMongoWrapper {
	builder.query = query
	return builder
}

func (builder *WorkloadMetricQueryMongoWrapper) GetQuery() *WorkloadMetricQueryParam {
	return builder.query
}