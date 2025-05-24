package result

import (
	"duolingo/lib/metric"
	"duolingo/lib/metric/downsampling"
	"time"
)

type WorkloadMetricQueryResult struct {
	MetricTarget string `json:"metric_target" bson:"metric_target"`
	MetricName string `json:"metric_name" bson:"metric_name"`
	Snapshots []*metric.Snapshot `json:"snapshots" bson:"snapshots"`
	ReducedSnapshots map[string][]*metric.Snapshot `json:"reduced_snapshots" bson:"reduced_snapshots"`
}

func (result *WorkloadMetricQueryResult) Downsampling(workloadStart time.Time, reductionStep int64, strategies map[string]downsampling.DownsamplingStrategy) error {
	reducer := new(downsampling.SnapshotReducer).
					WithStartTime(workloadStart).
					WithReductionStep(int64(reductionStep)).
					WithSnapshots(result.Snapshots)
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
