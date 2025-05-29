package param

import "duolingo/lib/metric/reduction"

type WorkloadMetricReduction struct {
	ReductionStep int64
	Stratergies   map[string]reduction.ReductionStrategy
}
