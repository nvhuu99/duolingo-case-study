package param

import "duolingo/constant"

type WorkloadMetricQueryFirebaseWrapper struct {
	query *WorkloadMetricQueryParam
}

func NewWorkloadFirebaseMetricQuery(traceId string) *WorkloadMetricQueryFirebaseWrapper {
	builder := new(WorkloadMetricQueryFirebaseWrapper)
	builder.query = NewWorkloadMetricQueryParam(traceId).SetMetricTarget(constant.METRIC_TARGET_FIREBASE)
	return builder
}

func (builder *WorkloadMetricQueryFirebaseWrapper) AddMetricNames(metricNames... string) *WorkloadMetricQueryFirebaseWrapper {
	builder.query.AddMetricNames(metricNames...)
	return builder
}

func (builder *WorkloadMetricQueryFirebaseWrapper) SetServiceName(name string) *WorkloadMetricQueryFirebaseWrapper {
	builder.query.SetServiceName(name)
	return builder
}

func (builder *WorkloadMetricQueryFirebaseWrapper) SetServiceOperation(opt string) *WorkloadMetricQueryFirebaseWrapper {
	builder.query.SetServiceOperation(opt)
	return builder
}

func (builder *WorkloadMetricQueryFirebaseWrapper) SetQuery(query *WorkloadMetricQueryParam) *WorkloadMetricQueryFirebaseWrapper {
	builder.query = query
	return builder
}

func (builder *WorkloadMetricQueryFirebaseWrapper) GetQuery() *WorkloadMetricQueryParam {
	return builder.query
}