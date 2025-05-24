package downsampling

import "duolingo/lib/metric"

type DownsamplingStrategy interface {
	UseSource(SnapshotReduction)
	Make(reductionStep int64, dp []*metric.Snapshot) (*metric.Snapshot, error)
}
