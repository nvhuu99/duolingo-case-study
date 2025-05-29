package reduction

import (
	"duolingo/lib/metric"
	"errors"
	"time"
)

type Sum struct {
	source SnapshotReduction
}

func (ma *Sum) UseSource(src SnapshotReduction) {
	ma.source = src
}

func (ma *Sum) Make(reduction int64, snapshots []*metric.Snapshot) (*metric.Snapshot, error) {
	if len(snapshots) == 0 {
		return nil, errors.New("reduction is empty")
	}

	var sumValue float64
	var sumTimestamp int64

	for _, d := range snapshots {
		sumValue += d.Value
		sumTimestamp += d.Timestamp.UnixMilli()
	}
	avgTimestamp := sumTimestamp / int64(len(snapshots))

	avg := &metric.Snapshot{Value: sumValue, Timestamp: time.UnixMilli(avgTimestamp)}
	return avg, nil
}
