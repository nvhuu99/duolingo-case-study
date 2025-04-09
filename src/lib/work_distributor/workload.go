package work_distributor

import "math"

type Workload struct {
	Name             string `json:"name"`
	NumOfUnits       int    `json:"num_of_units"`
	DistributionSize int    `json:"distribution_size"`
}

func (workload *Workload) NumOfAssignments() int {
	return int(math.Ceil(float64(workload.NumOfUnits) / float64(workload.DistributionSize)))
}

func (workload *Workload) ValidAttributes() bool {
	return workload.NumOfUnits > 0 &&
		workload.DistributionSize > 0 &&
		workload.NumOfUnits >= workload.DistributionSize
}
