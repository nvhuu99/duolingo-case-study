package result

import (
	"duolingo/constant"
	"duolingo/lib/metric/reduction"
	"slices"
)

type WorkloadMetricQueryResultRabbitMQWrapper struct {
	queryResult *WorkloadMetricQueryResult
}

func NewWorkloadRabbitMQQueryResult(queryResult *WorkloadMetricQueryResult) *WorkloadMetricQueryResultRabbitMQWrapper {
	wrapper := new(WorkloadMetricQueryResultRabbitMQWrapper)
	wrapper.queryResult = queryResult
	return wrapper
}

func (wrapper *WorkloadMetricQueryResultRabbitMQWrapper) Reduce(workload *WorkloadMetadataResult, reductionStep int64, strategies map[string]reduction.ReductionStrategy) error {
	rm := wrapper.queryResult

	accumulation := []string{
		constant.METRIC_NAME_DELIVERED_RATE,
		constant.METRIC_NAME_PUBLISHED_RATE,
	}
	sumStrg := map[string]reduction.ReductionStrategy{"sum": new(reduction.Sum)}
	if slices.Contains(accumulation, rm.MetricName) {
		if err := rm.Reduce(workload, workload.Incr, sumStrg); err != nil {
			return err
		}
		rm.Snapshots = rm.ReducedSnapshots["sum"]
		delete(rm.ReducedSnapshots, "sum")
	}

	if len(strategies) > 0 {
		if err := rm.Reduce(workload, reductionStep, strategies); err != nil {
			return err
		}
	}

	return nil
}
