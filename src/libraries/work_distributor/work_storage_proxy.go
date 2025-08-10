package work_distributor

import (
	"context"
	"errors"
)

var (
	ErrWorkloadNotExists = errors.New("workload not exists")
)

type WorkStorageProxy interface {
	SaveWorkload(ctx context.Context, w *Workload) error
	GetWorkload(ctx context.Context,workloadId string) (*Workload, error)
	GetAndUpdateWorkload(ctx context.Context,workloadId string, modifier func(*Workload) error) error
	DeleteWorkloadAndAssignments(ctx context.Context,workloadId string) error
	PushAssignmentToQueue(ctx context.Context,assignment *Assignment) error
	PopAssignmentFromQueue(ctx context.Context,workloadId string) (*Assignment, error)
}
