package result

import (
	"duolingo/constant"
	"duolingo/lib/metric/reduction"
	"slices"
)

type WorkloadMetricQueryResultRedisWrapper struct {
	queryResult *WorkloadMetricQueryResult
}

func NewWorkloadRedisQueryResult(queryResult *WorkloadMetricQueryResult) *WorkloadMetricQueryResultRedisWrapper {
	wrapper := new(WorkloadMetricQueryResultRedisWrapper)
	wrapper.queryResult = queryResult
	return wrapper
}

func (wrapper *WorkloadMetricQueryResultRedisWrapper) Reduce(workload *WorkloadMetadataResult, reductionStep int64, strategies map[string]reduction.ReductionStrategy) error {
	rm := wrapper.queryResult

	accumulation := []string{
		constant.METRIC_NAME_REDIS_CMD_RATE,
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
