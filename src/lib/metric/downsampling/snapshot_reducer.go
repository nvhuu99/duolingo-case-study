package downsampling

import (
	"duolingo/lib/metric"
	"errors"
	"sort"
	"time"
)

type SnapshotReducer struct {
	reductionStep int64
	maxReductionStep int64
	startTimeMs   int64
	strategy      DownsamplingStrategy
	reduced       map[int64][]*metric.Snapshot
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

/* Downsampling setup and execution methods */

func (ds *SnapshotReducer) WithReductionStep(step int64) *SnapshotReducer {
	if step > 0 {
		ds.reductionStep = step
	}
	return ds
}

func (ds *SnapshotReducer) WithStartTime(start time.Time) *SnapshotReducer {
	ds.startTimeMs = start.UnixMilli()
	return ds
}

func (ds *SnapshotReducer) WithStrategy(stg DownsamplingStrategy) *SnapshotReducer {
	ds.strategy = stg
	stg.UseSource(ds)
	return ds
}

func (ds *SnapshotReducer) WithSnapshots(datapoints []*metric.Snapshot) *SnapshotReducer {
	reduced := make(map[int64][]*metric.Snapshot)
	for _, snp := range datapoints {
		// calculate the reduction value based on snapshot's timestamp
		timeDiff := snp.Timestamp.UnixMilli() - ds.startTimeMs
		rd := ds.reductionStep * ((timeDiff + ds.reductionStep - 1) / ds.reductionStep)
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

// Downsampling the reduced snapshots using the DownsamplingStrategy
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
