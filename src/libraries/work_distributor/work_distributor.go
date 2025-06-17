package work_distributor

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWorkloadHasAlreadyFulfilled = errors.New("workload has already fulfilled as all assignments commited")
)

type WorkDistributor struct {
	proxy WorkStorageProxy

	unitsPerAssignment uint64
}

func NewWorkDistributor(proxy WorkStorageProxy, distributionSize uint64) *WorkDistributor {
	return &WorkDistributor{
		proxy:              proxy,
		unitsPerAssignment: distributionSize,
	}
}

func (dist *WorkDistributor) GetDistributionSize() uint64 {
	return dist.unitsPerAssignment
}

func (dist *WorkDistributor) CreateWorkload(totalWorkUnits uint64) (*Workload, error) {
	workload, validateErr := NewWorkload(
		uuid.NewString(),
		totalWorkUnits,
		dist.unitsPerAssignment,
	)
	if validateErr != nil {
		return nil, validateErr
	}
	// Create assignments, and push assignments to the queue
	var total = workload.GetExpectTotalAssignments()
	for i := range total {
		start := i*dist.unitsPerAssignment + 1
		end := start + dist.unitsPerAssignment - 1
		assignment, validationErr := NewAssignment(
			uuid.NewString(),
			workload.Id,
			start,
			end,
		)
		if validationErr != nil {
			return nil, validationErr
		}
		pushErr := dist.proxy.PushAssignmentToQueue(assignment)
		if pushErr != nil {
			return nil, pushErr
		}
	}
	// Save workload only after queuing all assignments
	saveErr := dist.proxy.SaveWorkload(workload)
	if saveErr != nil {
		return nil, saveErr
	}
	return workload, nil
}

func (dist *WorkDistributor) GetWorkload(workloadId string) (*Workload, error) {
	return dist.proxy.GetWorkload(workloadId)
}

func (dist *WorkDistributor) HasWorkloadFulfilled(workloadId string) (bool, error) {
	workload, err := dist.proxy.GetWorkload(workloadId)
	if err != nil {
		return false, err
	}
	return workload.HasWorkloadFulfilled(), nil
}

func (dist *WorkDistributor) Assign(workloadId string) (*Assignment, error) {
	isFullfilled, err := dist.HasWorkloadFulfilled(workloadId)
	if err != nil {
		return nil, err
	}
	if isFullfilled {
		return nil, ErrWorkloadHasAlreadyFulfilled
	}
	return dist.proxy.PopAssignmentFromQueue(workloadId)
}

func (dist *WorkDistributor) WaitForAssignment(
	waitCtx context.Context,
	retryWait time.Duration,
	workloadId string,
) (
	*Assignment,
	error,
) {
	for {
		select {
		case <-waitCtx.Done():
			return nil, errors.New("stop waiting for assignment due to context canceled")
		default:
			assignment, err := dist.Assign(workloadId)
			// the queue is empty, but the workload not yet fulfilled
			if assignment == nil && err == nil {
				time.Sleep(retryWait)
				continue
			}
			// either the workload has fulfilled, or operational error
			if err != nil {
				return nil, err
			}
			return assignment, nil
		}
	}
}

func (dist *WorkDistributor) HandleAssignment(
	assignment *Assignment,
	closure func() error,
) error {
	if handleErr := closure(); handleErr != nil {
		return dist.Rollback(assignment)
	}
	return dist.Commit(assignment)
}

func (dist *WorkDistributor) Commit(assignment *Assignment) error {
	return dist.proxy.GetAndUpdateWorkload(assignment.WorkloadId, func(w *Workload) error {
		return w.IncreaseTotalCommittedAssignments()
	})
}

func (dist *WorkDistributor) Rollback(assignment *Assignment) error {
	pushErr := dist.proxy.PushAssignmentToQueue(assignment)
	return pushErr
}

func (dist *WorkDistributor) CommitProgress(assignment *Assignment, newProgres uint64) error {
	assignment.Progress = newProgres
	if assignment.IsCompleted() {
		return dist.Commit(assignment)
	}
	return dist.proxy.PushAssignmentToQueue(assignment)
}

func (dist *WorkDistributor) DeleteWorkloadAndAssignments(workloadId string) error {
	return dist.proxy.DeleteWorkloadAndAssignments(workloadId)
}
