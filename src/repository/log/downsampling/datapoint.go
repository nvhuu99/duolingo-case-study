package downsampling

import "time"

type DataPoint struct {
	timestamp time.Time
	value     float64
}

func NewDataPoint(timestamp time.Time, value float64) *DataPoint {
	return &DataPoint{timestamp, value}
}

func (dp *DataPoint) GetValue() float64 {
	return dp.value
}

func (dp *DataPoint) GetTimestamp() time.Time {
	return dp.timestamp
}