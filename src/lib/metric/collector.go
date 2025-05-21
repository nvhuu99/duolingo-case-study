package metric

type Collector interface {
	Capture()
	Collect() []*DataPoint
}
