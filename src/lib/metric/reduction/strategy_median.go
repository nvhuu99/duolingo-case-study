package reduction

import (
	"duolingo/lib/metric"
	"errors"
	"time"
)

type Median struct {
	source SnapshotReduction
}

func (ma *Median) UseSource(src SnapshotReduction) {
	ma.source = src
}

func (ma *Median) Make(reduction int64, snapshots []*metric.Snapshot) (*metric.Snapshot, error) {
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

	avg := &metric.Snapshot{
		Value: avgValue, 
		Timestamp: time.UnixMilli(avgTimestamp),
		StartTimeOffset: reduction,
	}
	return avg, nil
}
