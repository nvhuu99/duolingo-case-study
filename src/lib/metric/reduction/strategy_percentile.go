package reduction

import (
	"duolingo/lib/metric"
	"errors"
	"sort"
	"time"
)

type PercentileStrategy struct {
	percentile float64
	source     SnapshotReduction
}

func NewPercentileStrategy(p float64) *PercentileStrategy {
	return &PercentileStrategy{percentile: p}
}

func (ps *PercentileStrategy) UseSource(src SnapshotReduction) {
	ps.source = src
}

func (ps *PercentileStrategy) Make(reduction int64, dp []*metric.Snapshot) (*metric.Snapshot, error) {
	if len(dp) == 0 {
		return nil, errors.New("reduction is empty")
	}

	// Extract and sort values
	values := make([]float64, len(dp))
	timestamps := make([]int64, len(dp))
	for i, point := range dp {
		values[i] = point.Value
		timestamps[i] = point.Timestamp.UnixMilli()
	}

	sort.Float64s(values)

	// Calculate percentile index
	pos := ps.percentile / 100.0 * float64(len(values)-1)
	lower := int(pos)
	upper := lower + 1
	weight := pos - float64(lower)

	var percentileValue float64
	if upper < len(values) {
		percentileValue = values[lower]*(1-weight) + values[upper]*weight
	} else {
		percentileValue = values[lower]
	}

	// Use median timestamp as a simple approximation
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
	medianTs := timestamps[len(timestamps)/2]

	result := &metric.Snapshot{
		Value: percentileValue, 
		Timestamp: time.UnixMilli(medianTs),
		StartTimeOffset: reduction,
	}
	return result, nil
}
