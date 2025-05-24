package param

import "duolingo/lib/metric/downsampling"

type WorkloadMetricDownsampling struct {
	ReductionStep int64
	Stratergies   map[string]downsampling.DownsamplingStrategy
}
