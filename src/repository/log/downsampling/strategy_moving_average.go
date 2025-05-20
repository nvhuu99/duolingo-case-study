package downsampling

import (
	"errors"
	"time"
)

type MovingAverage struct {
	source ReducedDataPoints
}

func (ma *MovingAverage) UseSource(src ReducedDataPoints) {
	ma.source = src
}

func (ma *MovingAverage) Make(reduction int64, dp []*DataPoint) (*DataPoint, error) {
	if len(dp) == 0 {
		return nil, errors.New("reduction is empty")
	}

	var sumValue float64
	var sumTimestamp int64

	for _, d := range dp {
		sumValue += d.GetValue()
		sumTimestamp += d.GetTimestamp().UnixMilli()
	}

	avgValue := sumValue / float64(len(dp))
	avgTimestamp := sumTimestamp / int64(len(dp))

	return NewDataPoint(time.UnixMilli(avgTimestamp), avgValue), nil
}