package downsampling

import (
	"duolingo/lib/metric"
	"errors"
	"time"
)

type MovingAverage struct {
	source SnapshotReduction
}

func (ma *MovingAverage) UseSource(src SnapshotReduction) {
	ma.source = src
}

func (ma *MovingAverage) Make(reduction int64, snapshots []*metric.Snapshot) (*metric.Snapshot, error) {
	if len(snapshots) == 0 {
		return nil, errors.New("reduction is empty")
	}

	var sumValue float64
	var sumTimestamp int64

	for _, d := range snapshots {
		sumValue += d.Value
		sumTimestamp += d.Timestamp.UnixMilli()
	}
	avgValue := sumValue / float64(len(snapshots))
	avgTimestamp := sumTimestamp / int64(len(snapshots))

	avg := &metric.Snapshot{Value: avgValue, Timestamp: time.UnixMilli(avgTimestamp)}
	return avg, nil
}
