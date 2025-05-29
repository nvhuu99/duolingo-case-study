package param

type WorkloadMetricQueryParam struct {
	Filters struct {
		TraceId          string
		InstanceIds      []string
		ServiceName      string
		ServiceOperation string
		Metadata map[string]string
	}
	MetricTarget string
	MetricNames []string
}

func NewWorkloadMetricQueryParam(traceId string) *WorkloadMetricQueryParam {
	params := new(WorkloadMetricQueryParam)
	params.Filters.TraceId = traceId
	return params
}

func (params *WorkloadMetricQueryParam) SetServiceName(name string) *WorkloadMetricQueryParam {
	params.Filters.ServiceName = name
	return params
}

func (params *WorkloadMetricQueryParam) SetServiceOperation(opt string) *WorkloadMetricQueryParam {
	params.Filters.ServiceOperation = opt
	return params
}

func (params *WorkloadMetricQueryParam) SetServiceInstanceIds(instanceIds []string) *WorkloadMetricQueryParam {
	params.Filters.InstanceIds = instanceIds
	return params
}

func (params *WorkloadMetricQueryParam) SetMetricTarget(target string) *WorkloadMetricQueryParam {
	params.MetricTarget = target
	return params
}

func (params *WorkloadMetricQueryParam) AddMetricNames(metricNames... string) *WorkloadMetricQueryParam {
	params.MetricNames = append(params.MetricNames, metricNames...)
	return params
}

func (params *WorkloadMetricQueryParam) AddMetadata(key string, value string) *WorkloadMetricQueryParam {
	if params.Filters.Metadata == nil {
		params.Filters.Metadata = make(map[string]string)
	} 
	params.Filters.Metadata[key] = value
	return params
}