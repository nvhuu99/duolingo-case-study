package work_distributor

import (
	"errors"
	"time"
)

var (
	ErrUnexpectedWorkloadTotalCommittedAssignments = errors.New("workload total committed exceeds expectation")
)

type Workload struct {
	Id                        string `json:"id"`
	TotalWorkUnits            uint64 `json:"total_units"`
	TotalUnitsPerAssignment   uint64 `json:"dist_size"`
	TotalCommittedAssignments uint64 `json:"total_commited"`

	CreatedAt time.Time
}

func NewWorkload(id string, totalUnits uint64, unitsPerAssignment uint64) (*Workload, error) {
	workload := &Workload{
		Id:                        id,
		TotalWorkUnits:            totalUnits,
		TotalUnitsPerAssignment:   unitsPerAssignment,
		TotalCommittedAssignments: 0,
		CreatedAt:                 time.Now(),
	}

	if validationErr := workload.Validate(); validationErr != nil {
		return nil, validationErr
	}

	return workload, nil
}

func (w *Workload) Validate() error {
	if w.Id == "" ||
		w.TotalWorkUnits == 0 ||
		w.TotalUnitsPerAssignment == 0 {
		return errors.New("invalid workload's parameters")
	}
	return nil
}

func (w *Workload) GetExpectTotalAssignments() uint64 {
	size := w.TotalUnitsPerAssignment
	return (w.TotalWorkUnits + size - 1) / size // round up division
}

func (w *Workload) HasWorkloadFulfilled() bool {
	return w.TotalCommittedAssignments == w.GetExpectTotalAssignments()
}

func (w *Workload) IncreaseTotalCommittedAssignments() error {
	if w.TotalCommittedAssignments == w.GetExpectTotalAssignments() {
		return ErrUnexpectedWorkloadTotalCommittedAssignments
	}
	w.TotalCommittedAssignments++
	return nil
}

func (w *Workload) Equal(target *Workload) bool {
	return target != nil && target.Id == w.Id
}
