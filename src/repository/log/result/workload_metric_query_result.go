package result

import (
	"duolingo/lib/metric"
	"duolingo/lib/metric/reduction"
)

type WorkloadMetricQueryResult struct {
	MetricTarget     string                        `json:"metric_target" bson:"metric_target"`
	MetricName       string                        `json:"metric_name" bson:"metric_name"`
	Snapshots        []*metric.Snapshot            `json:"snapshots" bson:"snapshots"`
	ReducedSnapshots map[string][]*metric.Snapshot `json:"reduced_snapshots" bson:"reduced_snapshots"`
}

func (result *WorkloadMetricQueryResult) Reduce(workload *WorkloadMetadataResult, reductionStep int64, strategies map[string]reduction.ReductionStrategy) error {
	reducer := new(reduction.SnapshotReducer).
		WithStartTime(workload.StartTime).
		WithSnapshots(result.Snapshots, reductionStep)

	result.ReducedSnapshots = make(map[string][]*metric.Snapshot)
	
	for name, strategy := range strategies {
		reducedWithStrg, err := reducer.WithStrategy(strategy).Result()
		if err != nil {
			return err
		}
		result.ReducedSnapshots[name] = reducedWithStrg
	}

	return nil
}
