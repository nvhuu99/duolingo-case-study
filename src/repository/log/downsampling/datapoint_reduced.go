package downsampling

type ReducedDataPoints interface {
	GetDataPoint(reductionStep int64, dpIndex int) *DataPoint
	GetReducedDataPoints(reductionStep int64) []*DataPoint
	GetReductionStep() int64
	TotalReductions() int64
	NextReduction(current int64) (int64, error)
	PreviousReduction(current int64) (int64, error)
}
