package reduction

import "duolingo/lib/metric"

type ReductionStrategy interface {
	UseSource(SnapshotReduction)
	Make(reductionStep int64, dp []*metric.Snapshot) (*metric.Snapshot, error)
}
