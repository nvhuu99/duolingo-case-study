package reduction

import (
	"duolingo/lib/metric"
	"errors"
	"time"
)

type Max struct {
	source SnapshotReduction
}

func (ma *Max) UseSource(src SnapshotReduction) {
	ma.source = src
}

func (ma *Max) Make(reduction int64, snapshots []*metric.Snapshot) (*metric.Snapshot, error) {
	if len(snapshots) == 0 {
		return nil, errors.New("reduction is empty")
	}

	var maxVal float64 = snapshots[0].Value
	var sumTimestamp int64

	for _, d := range snapshots {
		if d.Value > maxVal {
			maxVal = d.Value
		}
		sumTimestamp += d.Timestamp.UnixMilli()
	}
	avgTimestamp := sumTimestamp / int64(len(snapshots))

	avg := &metric.Snapshot{Value: maxVal, Timestamp: time.UnixMilli(avgTimestamp)}
	return avg, nil
}
