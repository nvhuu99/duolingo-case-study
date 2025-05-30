package result

import (
	"duolingo/lib/metric/reduction"
)

type WorkloadMetricQueryResultFirebaseWrapper struct {
	queryResult *WorkloadMetricQueryResult
}

func NewWorkloadFirebaseQueryResult(queryResult *WorkloadMetricQueryResult) *WorkloadMetricQueryResultFirebaseWrapper {
	wrapper := new(WorkloadMetricQueryResultFirebaseWrapper)
	wrapper.queryResult = queryResult
	return wrapper
}

func (wrapper *WorkloadMetricQueryResultFirebaseWrapper) Reduce(workload *WorkloadMetadataResult, reductionStep int64, strategies map[string]reduction.ReductionStrategy) error {
	rm := wrapper.queryResult

	if len(strategies) > 0 {
		if err := rm.Reduce(workload, reductionStep, strategies); err != nil {
			return err
		}
	}

	return nil
}
