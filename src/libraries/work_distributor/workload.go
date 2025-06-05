package work_distributor

import "errors"

type Workload struct {
	id                       string
	totalWorkloadUnits       uint64
	totalUnitsPerAssignment  uint64
	totalPendingAssignments  uint64
	totalCommittedAssignment uint64
}

func NewWorkload(id string, totalUnits uint64, unitsPerAssignment uint64) (*Workload, error) {
	if id == "" || totalUnits == 0 || unitsPerAssignment == 0 {
		return nil, errors.New("invalid workload's parameters")
	}
	workload := &Workload{
		id:                      id,
		totalWorkloadUnits:      totalUnits,
		totalUnitsPerAssignment: unitsPerAssignment,
	}
	return workload, nil
}

func (w *Workload) GetId() string {
	return w.id
}

func (w *Workload) GetDistributionSize() uint64 {
	return w.totalUnitsPerAssignment
}

func (w *Workload) GetTotalWorkloadUnits() uint64 {
	return w.totalWorkloadUnits
}

// func (w *Workload) GetTotalPendingAssignments() uint64 {
// 	return w.totalPendingAssignments
// }

// func (w *Workload) GetTotalCommitedAssignments() uint64 {
// 	return w.totalCommittedAssignment
// }
