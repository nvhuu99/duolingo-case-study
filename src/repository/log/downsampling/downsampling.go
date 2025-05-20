package downsampling

import (
	"errors"
	"sort"
	"time"
)

type Downsampling struct {
	reductionStep int64
	startTimeMs int64
	strategy DownsamplingStrategy
	reduced map[int64][]*DataPoint
}

/* Implement the ReducedDataPoints interface*/

func (ds *Downsampling) GetDataPoint(reduction int64, dpIndex int) *DataPoint {
	if dp, exist := ds.reduced[reduction]; exist {
		if dpIndex < len(dp) {
			return dp[dpIndex]
		}
	}
	return nil
}

func (ds *Downsampling) GetReducedDataPoints(reduction int64) []*DataPoint {
	if dp, exist := ds.reduced[reduction]; exist {
		return dp
	}
	return nil
}

func (ds *Downsampling) GetReductionStep() int64 {
	return ds.reductionStep
}

func (ds *Downsampling) TotalReductions() int64 {
	return int64(len(ds.reduced))
}

func (ds *Downsampling) PreviousReduction(current int64) (int64, error) {
	prev := current - ds.GetReductionStep()
	if prev < 0 {
		return 0, errors.New("can not get previous reduction value, must be greater or equal zero")
	}
	return prev, nil
}

func (ds *Downsampling) NextReduction(current int64) (int64, error) {
	next := current + ds.GetReductionStep()
	max := int64(ds.TotalReductions()) * ds.GetReductionStep() 
	if next > max {
		return 0, errors.New("can not get next reduction value, max reduction value exceeded")
	}
	return next, nil
}

/* Downsampling setup and execution methods */

func (ds *Downsampling) WithReductionStep(step int64) *Downsampling {
	if step > 0 {
		ds.reductionStep = step
	}
	return ds
}

func (ds *Downsampling) WithStartTime(start time.Time) *Downsampling {
	ds.startTimeMs = start.UnixMilli()
	return ds
}

func (ds *Downsampling) WithStrategy(stg DownsamplingStrategy) *Downsampling {
	ds.strategy = stg
	stg.UseSource(ds)
	return ds
}

func (ds *Downsampling) WithDatapoints(datapoints []*DataPoint) *Downsampling {
	reduced := make(map[int64][]*DataPoint)
	for _, dp := range datapoints {
		// calculate the reduction value based on datapoint's timestamp
		timeDiff := dp.GetTimestamp().UnixMilli() - ds.startTimeMs
		rd := ds.reductionStep * ((timeDiff + ds.reductionStep - 1) / ds.reductionStep)
		// push the datapoint to the reduction group
		if _, exist := reduced[rd]; !exist {
			reduced[rd] = []*DataPoint{}
		}
		reduced[rd] = append(reduced[rd], dp)
	}
	ds.reduced = reduced

	return ds
}

// Downsampling the reduced datapoints using the DownsamplingStrategy
func (ds *Downsampling) Result() ([]*DataPoint, error) {
	if ds.TotalReductions() == 0 {
		return []*DataPoint{}, nil
	}
	
	result := []*DataPoint{}
	
	for rd, datapoints := range ds.reduced {
		if len(datapoints) > 0 {
			dp, err := ds.strategy.Make(rd, datapoints)
			if err != nil {
				return nil, err
			}
			result = append(result, dp)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].GetTimestamp().Before(result[j].GetTimestamp())
	})

	return result, nil
}