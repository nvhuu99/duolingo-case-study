package param

import "duolingo/constant"

type WorkloadMetricQueryRedisWrapper struct {
	query *WorkloadMetricQueryParam
}

func NewWorkloadRedisMetricQuery(traceId string) *WorkloadMetricQueryRedisWrapper {
	builder := new(WorkloadMetricQueryRedisWrapper)
	builder.query = NewWorkloadMetricQueryParam(traceId).SetMetricTarget(constant.METRIC_TARGET_REDIS)
	return builder
}

func (builder *WorkloadMetricQueryRedisWrapper) AddMetricNames(metricNames... string) *WorkloadMetricQueryRedisWrapper {
	builder.query.AddMetricNames(metricNames...)
	return builder
}

func (builder *WorkloadMetricQueryRedisWrapper) SetServiceName(name string) *WorkloadMetricQueryRedisWrapper {
	builder.query.SetServiceName(name)
	return builder
}

func (builder *WorkloadMetricQueryRedisWrapper) SetServiceOperation(opt string) *WorkloadMetricQueryRedisWrapper {
	builder.query.SetServiceOperation(opt)
	return builder
}

func (builder *WorkloadMetricQueryRedisWrapper) SetQuery(query *WorkloadMetricQueryParam) *WorkloadMetricQueryRedisWrapper {
	builder.query = query
	return builder
}

func (builder *WorkloadMetricQueryRedisWrapper) GetQuery() *WorkloadMetricQueryParam {
	return builder.query
}