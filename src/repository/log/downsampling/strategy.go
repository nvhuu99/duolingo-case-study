package downsampling

type DownsamplingStrategy interface {
	UseSource(ReducedDataPoints)
	Make(reductionStep int64, dp []*DataPoint) (*DataPoint, error)
}
