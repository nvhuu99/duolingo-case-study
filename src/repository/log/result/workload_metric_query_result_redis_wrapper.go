package result

// import (
// 	"duolingo/constant"
// 	"duolingo/lib/metric/reduction"
// 	"time"
// )

// type WorkloadMetricQueryResultRedisWrapper struct {
// 	queryResult *WorkloadMetricQueryResult
// }

// func NewWorkloadRedisQueryResult(queryResult *WorkloadMetricQueryResult) *WorkloadMetricQueryResultRedisWrapper {
// 	wrapper := new(WorkloadMetricQueryResultRedisWrapper)
// 	wrapper.queryResult = queryResult
// 	return wrapper
// }

// func (wrapper *WorkloadMetricQueryResultRedisWrapper) Reduce(workloadStart time.Time, reductionStep int64, strategies map[string]reduction.ReductionStrategy) error {
// 	rm := wrapper.queryResult

// 	if rm.MetricName == constant.METRIC_NAME_REDIS_CMD_COUNT {
// 		sumStrg := map[string]reduction.ReductionStrategy{"sum": new(reduction.Sum)}
// 		if err := rm.Reduce(workloadStart, reduction.REDUCTION_BY_SNAPSHOTS_INCR, sumStrg); err != nil {
// 			return err
// 		}
// 		rm.Snapshots = rm.ReducedSnapshots["sum"]
// 	}

// 	if len(strategies) > 0 {
// 		if err := rm.Reduce(workloadStart, reductionStep, strategies); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
