package downsampling

import "duolingo/lib/metric"

type SnapshotReduction interface {
	GetSnapshot(reductionStep int64, snapshotIdx int) *metric.Snapshot
	GetSnapshots(reductionStep int64) []*metric.Snapshot
	GetReductionStep() int64
	TotalReductions() int64
	NextReduction(current int64) (int64, error)
	PreviousReduction(current int64) (int64, error)
}
