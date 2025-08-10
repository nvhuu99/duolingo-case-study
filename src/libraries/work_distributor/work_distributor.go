package work_distributor

import (
	"context"
	"errors"
	"time"

	events "duolingo/libraries/events/facade"

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

	evt := events.Start(ctx, "work_dist.create_workload", nil)

	workload, validateErr := NewWorkload(
		uuid.NewString(),
		totalWorkUnits,
		dist.unitsPerAssignment,
	)
	if validateErr != nil {
		events.Failed(evt, validateErr, nil)
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
			events.Failed(evt, validationErr, nil)
			return nil, validationErr
		}
		pushErr := dist.proxy.PushAssignmentToQueue(evt.Context(), assignment)
		if pushErr != nil {
			events.Failed(evt, pushErr, nil)
			return nil, pushErr
		}
	}
	// Save workload only after queuing all assignments
	saveErr := dist.proxy.SaveWorkload(ctx, workload)
	if saveErr != nil {
		events.Failed(evt, saveErr, nil)
		return nil, saveErr
	}

	events.Succeeded(evt, nil)

	return workload, nil
}

func (dist *WorkDistributor) GetWorkload(
	ctx context.Context,
	workloadId string,
) (*Workload, error) {
	var workload *Workload
	var err error

	evt := events.Start(ctx, "work_dist.get_workload", nil)
	defer events.End(evt, true, err, nil)

	workload, err = dist.proxy.GetWorkload(evt.Context(), workloadId)

	return workload, err
}

func (dist *WorkDistributor) HasWorkloadFulfilled(
	ctx context.Context,
	workloadId string,
) (bool, error) {
	var workload *Workload
	var err error

	evt := events.Start(ctx, "work_dist.has_workload_fulfilled", nil)
	defer events.End(evt, true, err, nil)

	workload, err = dist.proxy.GetWorkload(ctx, workloadId)
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
	var assignment *Assignment
	var err error

	evt := events.Start(waitCtx, "work_dist.wait_for_assignment", nil)
	defer func() {
		if err != nil && err != ErrWorkloadHasAlreadyFulfilled {
			events.End(evt, true, err, nil)
		} else {
			events.Succeeded(evt, nil)
		}
	}()

	for {
		select {
		case <-waitCtx.Done():
			err = errors.New("stop waiting for assignment due to context canceled")
			return nil, err
		default:
			assignment, err = dist.Assign(evt.Context(), workloadId)
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
	var err error

	evt := events.Start(ctx, "work_dist.handle_assignment", nil)
	defer events.End(evt, true, err, nil)

	if err = handler(evt.Context()); err != nil {
		dist.Rollback(evt.Context(), assignment)
	} else {
		err = dist.Commit(evt.Context(), assignment)
	}

	return err
}

func (dist *WorkDistributor) Commit(ctx context.Context, assignment *Assignment) error {
	var err error

	evt := events.Start(ctx, "work_dist.commit", nil)
	defer events.End(evt, true, err, nil)

	err = dist.proxy.GetAndUpdateWorkload(evt.Context(), assignment.WorkloadId, func(
		w *Workload,
	) error {
		return w.IncreaseTotalCommittedAssignments()
	})

	return err
}

func (dist *WorkDistributor) Rollback(ctx context.Context, assignment *Assignment) error {
	var err error

	evt := events.Start(ctx, "work_dist.rollback", nil)
	defer events.End(evt, true, err, nil)

	err = dist.proxy.PushAssignmentToQueue(ctx, assignment)

	return err
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
