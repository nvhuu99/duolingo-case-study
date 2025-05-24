package param

type MetricGroup struct {
	MetricTarget string
	MetricName   string
}

type WorkloadMetricQuery struct {
	Filters struct {
		TraceId          string
		InstanceIds      []string
		ServiceName      string
		ServiceOperation string
	}

	MetricGroups []*MetricGroup
}

func WorkloadMetricQueryParams(traceId string) *WorkloadMetricQuery {
	params := new(WorkloadMetricQuery)
	params.Filters.TraceId = traceId
	return params
}

func (params *WorkloadMetricQuery) SetServiceName(name string) *WorkloadMetricQuery {
	params.Filters.ServiceName = name
	return params
}

func (params *WorkloadMetricQuery) SetServiceOperation(opt string) *WorkloadMetricQuery {
	params.Filters.ServiceOperation = opt
	return params
}

func (params *WorkloadMetricQuery) SetServiceInstanceIds(instanceIds []string) *WorkloadMetricQuery {
	params.Filters.InstanceIds = instanceIds
	return params
}

func (params *WorkloadMetricQuery) AddMetricGroup(metricTarget string, metricName string) *WorkloadMetricQuery {
	params.MetricGroups = append(params.MetricGroups, &MetricGroup{ metricTarget, metricName })
	return params
}
