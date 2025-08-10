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

	unitsPerAssignment int64
}

func NewWorkDistributor(proxy WorkStorageProxy, distributionSize int64) *WorkDistributor {
	return &WorkDistributor{
		proxy:              proxy,
		unitsPerAssignment: distributionSize,
	}
}

func (dist *WorkDistributor) GetDistributionSize() int64 {
	return dist.unitsPerAssignment
}

func (dist *WorkDistributor) CreateWorkload(
	ctx context.Context,
	totalWorkUnits int64,
) (*Workload, error) {
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
		pushErr := dist.proxy.PushAssignmentToQueue(ctx, assignment)
		if pushErr != nil {
			return nil, pushErr
		}
	}
	// Save workload only after queuing all assignments
	saveErr := dist.proxy.SaveWorkload(ctx, workload)
	if saveErr != nil {
		return nil, saveErr
	}
	return workload, nil
}

func (dist *WorkDistributor) GetWorkload(
	ctx context.Context,
	workloadId string,
) (*Workload, error) {
	return dist.proxy.GetWorkload(ctx, workloadId)
}

func (dist *WorkDistributor) HasWorkloadFulfilled(
	ctx context.Context,
	workloadId string,
) (bool, error) {
	workload, err := dist.proxy.GetWorkload(ctx, workloadId)
	if err != nil {
		return false, err
	}
	return workload.HasWorkloadFulfilled(), nil
}

func (dist *WorkDistributor) Assign(ctx context.Context, workloadId string) (*Assignment, error) {
	isFullfilled, err := dist.HasWorkloadFulfilled(ctx, workloadId)
	if err != nil {
		return nil, err
	}
	if isFullfilled {
		return nil, ErrWorkloadHasAlreadyFulfilled
	}
	return dist.proxy.PopAssignmentFromQueue(ctx, workloadId)
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
			assignment, err := dist.Assign(waitCtx, workloadId)
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
	ctx context.Context,
	assignment *Assignment,
	handler func(assignmentCtx context.Context) error,
) error {
	if handleErr := handler(ctx); handleErr != nil {
		return dist.Rollback(ctx, assignment)
	}
	return dist.Commit(ctx, assignment)
}

func (dist *WorkDistributor) Commit(ctx context.Context, assignment *Assignment) error {
	return dist.proxy.GetAndUpdateWorkload(ctx, assignment.WorkloadId, func(w *Workload) error {
		return w.IncreaseTotalCommittedAssignments()
	})
}

func (dist *WorkDistributor) Rollback(ctx context.Context, assignment *Assignment) error {
	pushErr := dist.proxy.PushAssignmentToQueue(ctx, assignment)
	return pushErr
}

func (dist *WorkDistributor) CommitProgress(
	ctx context.Context,
	assignment *Assignment,
	newProgres int64,
) error {
	assignment.Progress = newProgres
	if assignment.IsCompleted() {
		return dist.Commit(ctx, assignment)
	}
	return dist.proxy.PushAssignmentToQueue(ctx, assignment)
}

func (dist *WorkDistributor) DeleteWorkloadAndAssignments(
	ctx context.Context,
	workloadId string,
) error {
	return dist.proxy.DeleteWorkloadAndAssignments(ctx, workloadId)
}
