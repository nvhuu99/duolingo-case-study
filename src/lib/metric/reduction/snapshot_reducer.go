package reduction

import (
	"duolingo/lib/metric"
	"errors"
	"sort"
	"time"
)

type SnapshotReducer struct {
	reductionStep    int64
	maxReductionStep int64
	startTimeMs      int64
	strategy         ReductionStrategy
	reduced          map[int64][]*metric.Snapshot
}

/* Implement the SnapshotReduction interface*/

func (ds *SnapshotReducer) GetSnapshot(reduction int64, dpIndex int) *metric.Snapshot {
	if dp, exist := ds.reduced[reduction]; exist {
		if dpIndex < len(dp) {
			return dp[dpIndex]
		}
	}
	return nil
}

func (ds *SnapshotReducer) GetSnapshots(reduction int64) []*metric.Snapshot {
	if dp, exist := ds.reduced[reduction]; exist {
		return dp
	}
	return nil
}

func (ds *SnapshotReducer) GetReductionStep() int64 {
	return ds.reductionStep
}

func (ds *SnapshotReducer) TotalReductions() int64 {
	return int64(len(ds.reduced))
}

func (ds *SnapshotReducer) PreviousReduction(current int64) (int64, error) {
	prev := current - ds.GetReductionStep()
	if prev < 0 {
		return 0, errors.New("can not get previous reduction value, must be greater or equal zero")
	}
	return prev, nil
}

func (ds *SnapshotReducer) NextReduction(current int64) (int64, error) {
	next := current + ds.GetReductionStep()
	if next > ds.maxReductionStep {
		return 0, errors.New("can not get next reduction value, max reduction value exceeded")
	}
	return next, nil
}

/* Reduction setup and execution methods */

func (ds *SnapshotReducer) WithReductionStep() *SnapshotReducer {

	return ds
}

func (ds *SnapshotReducer) WithStartTime(start time.Time) *SnapshotReducer {
	ds.startTimeMs = start.UnixMilli()
	return ds
}

func (ds *SnapshotReducer) WithStrategy(stg ReductionStrategy) *SnapshotReducer {
	ds.strategy = stg
	stg.UseSource(ds)
	return ds
}

func (ds *SnapshotReducer) WithSnapshots(datapoints []*metric.Snapshot, reductionStep int64) *SnapshotReducer {
	if reductionStep == 0 {
		panic("reduction step can not be zero")
	}

	if ds.startTimeMs == (time.Time{}).UnixMilli() {
		panic("reduction workload start time can not be empty")
	}

	if len(datapoints) == 0 {
		return ds
	}
	step := reductionStep
	if step == REDUCTION_BY_SNAPSHOTS_INCR {
		sort.Slice(datapoints, func(i, j int) bool {
			return datapoints[i].Timestamp.Before(datapoints[j].Timestamp)
		})
		if len(datapoints) >= 2 {
			step = datapoints[1].Timestamp.UnixMilli() - datapoints[0].Timestamp.UnixMilli()
		} else {
			step = datapoints[0].Timestamp.UnixMilli() - ds.startTimeMs
		}
	}
	ds.reductionStep = step

	reduced := make(map[int64][]*metric.Snapshot)
	for _, snp := range datapoints {
		// calculate the reduction value based on snapshot's timestamp
		timeDiff := snp.Timestamp.UnixMilli() - ds.startTimeMs
		var rd int64
		if step != 0 {
			rd = step * ((timeDiff + step - 1) / step)
		}
		// push the snapshot to the reduction group
		reduced[rd] = append(reduced[rd], snp)
		// save the max reduction step
		if rd > ds.maxReductionStep {
			ds.maxReductionStep = rd
		}
	}
	for _, snapshots := range reduced {
		sort.Slice(snapshots, func(i, j int) bool {
			return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
		})
	}
	ds.reduced = reduced

	return ds
}

// Reduction the reduced snapshots using the ReductionStrategy
func (ds *SnapshotReducer) Result() ([]*metric.Snapshot, error) {
	if ds.TotalReductions() == 0 {
		return []*metric.Snapshot{}, nil
	}

	result := []*metric.Snapshot{}

	for rd, snapshots := range ds.reduced {
		if len(snapshots) > 0 {
			dp, err := ds.strategy.Make(rd, snapshots)
			if err != nil {
				return nil, err
			}
			// dp.Timestamp = time.UnixMilli(ds.startTimeMs).Add(time.Duration(rd) * time.Millisecond)
			result = append(result, dp)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result, nil
}
