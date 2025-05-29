package param

import (
	cnst "duolingo/constant"
)

type WorkloadMetricQueryRabbitMQWrapper struct {
	query *WorkloadMetricQueryParam
}

func NewWorkloadRabbitMQMetricQuery(traceId string) *WorkloadMetricQueryRabbitMQWrapper {
	builder := new(WorkloadMetricQueryRabbitMQWrapper)
	builder.query = NewWorkloadMetricQueryParam(traceId).SetMetricTarget(cnst.METRIC_TARGET_RABBITMQ)
	return builder
}

func (builder *WorkloadMetricQueryRabbitMQWrapper) AddMetricNames(metricName... string) *WorkloadMetricQueryRabbitMQWrapper {
	builder.query.AddMetricNames(metricName...)
	return builder
}

func (builder *WorkloadMetricQueryRabbitMQWrapper) SetServiceName(name string) *WorkloadMetricQueryRabbitMQWrapper {
	builder.query.SetServiceName(name)
	return builder
}

func (builder *WorkloadMetricQueryRabbitMQWrapper) SetServiceOperation(opt string) *WorkloadMetricQueryRabbitMQWrapper {
	builder.query.SetServiceOperation(opt)
	return builder
}

func (builder *WorkloadMetricQueryRabbitMQWrapper) SetMessageQueue(queue string) *WorkloadMetricQueryRabbitMQWrapper {
	builder.query.AddMetadata("queue", queue)
	return builder
}

func (builder *WorkloadMetricQueryRabbitMQWrapper) SetQuery(query *WorkloadMetricQueryParam) *WorkloadMetricQueryRabbitMQWrapper {
	builder.query = query
	return builder
}

func (builder *WorkloadMetricQueryRabbitMQWrapper) GetQuery() *WorkloadMetricQueryParam {
	return builder.query
}