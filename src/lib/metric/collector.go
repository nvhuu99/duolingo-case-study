package metric

type Collector interface {
	Capture() any
}
