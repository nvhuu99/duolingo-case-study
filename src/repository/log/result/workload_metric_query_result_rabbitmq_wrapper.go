package result

// import (
// 	"duolingo/constant"
// 	"duolingo/lib/metric/reduction"
// 	"slices"
// 	"time"
// )

// type WorkloadMetricQueryResultRabbitMQWrapper struct {
// 	queryResult *WorkloadMetricQueryResult
// }

// func NewWorkloadRabbitMQQueryResult(queryResult *WorkloadMetricQueryResult) *WorkloadMetricQueryResultRabbitMQWrapper {
// 	wrapper := new(WorkloadMetricQueryResultRabbitMQWrapper)
// 	wrapper.queryResult = queryResult
// 	return wrapper
// }

// func (wrapper *WorkloadMetricQueryResultRabbitMQWrapper) Reduce(workloadStart time.Time, reductionStep int64, strategies map[string]reduction.ReductionStrategy) error {
// 	rm := wrapper.queryResult

// 	accumulation := []string{
// 		constant.METRIC_NAME_QUEUE_DEPTH,
// 		constant.METRIC_NAME_DELIVERED,
// 		constant.METRIC_NAME_PUBLISHED,
// 	}
// 	sumStrg := map[string]reduction.ReductionStrategy{"sum": new(reduction.Sum)}
// 	if slices.Contains(accumulation, rm.MetricName) {
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
