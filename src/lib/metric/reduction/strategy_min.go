package reduction

import (
	"duolingo/lib/metric"
	"errors"
	"time"
)

type Min struct {
	source SnapshotReduction
}

func (ma *Min) UseSource(src SnapshotReduction) {
	ma.source = src
}

func (ma *Min) Make(reduction int64, snapshots []*metric.Snapshot) (*metric.Snapshot, error) {
	if len(snapshots) == 0 {
		return nil, errors.New("reduction is empty")
	}

	var minVal float64 = snapshots[0].Value
	var sumTimestamp int64

	for _, d := range snapshots {
		if d.Value < minVal {
			minVal = d.Value
		}
		sumTimestamp += d.Timestamp.UnixMilli()
	}
	avgTimestamp := sumTimestamp / int64(len(snapshots))

	avg := &metric.Snapshot{
		Value: minVal, 
		Timestamp: time.UnixMilli(avgTimestamp),
		StartTimeOffset: reduction,
	}
	return avg, nil
}
