package work_distributor

import (
	"errors"
)

var (
	ErrWorkloadNotExists = errors.New("workload not exists")
)

type WorkStorageProxy interface {
	SaveWorkload(w *Workload) error
	GetWorkload(workloadId string) (*Workload, error)
	GetAndUpdateWorkload(workloadId string, modifier func(*Workload) error) error
	DeleteWorkloadAndAssignments(workloadId string) error
	PushAssignmentToQueue(assignment *Assignment) error
	PopAssignmentFromQueue(workloadId string) (*Assignment, error)
}
