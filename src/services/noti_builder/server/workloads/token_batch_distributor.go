package workloads

import (
	"context"
	"duolingo/libraries/pub_sub"
	"duolingo/libraries/work_distributor"
	"duolingo/models"
	"duolingo/repositories/user_repository/external/services"
	"time"
)

type TokenBatchDistributor struct {
	*work_distributor.WorkDistributor

	jobPublisher  pub_sub.Publisher
	jobSubscriber pub_sub.Subscriber

	userService services.UserService
}

func NewTokenBatchDistributor(
	base *work_distributor.WorkDistributor,
	jobPublisher pub_sub.Publisher,
	jobSubscriber pub_sub.Subscriber,
	userService services.UserService,
) *TokenBatchDistributor {
	return &TokenBatchDistributor{
		WorkDistributor: base,
		jobPublisher:    jobPublisher,
		jobSubscriber:   jobSubscriber,
		userService:     userService,
	}
}

func (d *TokenBatchDistributor) CreateBatchJob(input *models.MessageInput) error {
	var err error
	var count uint64
	var workload *work_distributor.Workload
	if count, err = d.userService.CountDevicesForCampaign(input.Campaign); err == nil && count != 0 {
		if workload, err = d.CreateWorkload(count); err == nil {
			job := NewTokenBatchJob(workload.Id, input)
			err = d.jobPublisher.NotifyMainTopic(string(job.Encode()))
		}
	}
	return err
}

func (d *TokenBatchDistributor) ConsumingTokenBatches(
	ctx context.Context,
	batchConsumer func(input *models.MessageInput, devices []*models.UserDevice) error,
) error {
	return d.jobSubscriber.ConsumingMainTopic(ctx, func(str string) pub_sub.ConsumeAction {
		return d.acceptOrReject(
			d.startJobBatching(ctx, JobDecode([]byte(str)), batchConsumer))
	})
}

func (d *TokenBatchDistributor) startJobBatching(
	ctx context.Context,
	job *TokenBatchJob,
	closure func(input *models.MessageInput, devices []*models.UserDevice) error,
) error {
	if jobErr := job.Validate(); jobErr != nil {
		return jobErr
	}

	consumeCtx, consumeCancel := context.WithCancel(ctx)
	defer consumeCancel()

	var lastErr error
	var interval = 10 * time.Millisecond
	var jobId = job.JobId
	var assignment *work_distributor.Assignment
	for {
		select {
		case <-consumeCtx.Done():
			return nil
		default:
		}
		if lastErr != nil {
			return lastErr
		}
		if assignment, lastErr = d.WaitForAssignment(consumeCtx, interval, jobId); lastErr != nil {
			if lastErr == work_distributor.ErrWorkloadHasAlreadyFulfilled {
				return nil
			}
			continue
		}
		lastErr = d.HandleAssignment(assignment, func() error {
			devices, queryErr := d.userService.GetDevicesForCampaign(
				job.Message.Campaign,
				assignment.WorkStartAt()-1,                        // offset
				assignment.WorkEndAt()-assignment.WorkStartAt()+1, // limit
			)
			if queryErr != nil {
				return queryErr
			}
			return closure(job.Message, devices)
		})
	}
}

func (d *TokenBatchDistributor) acceptOrReject(err error) pub_sub.ConsumeAction {
	if err != nil {
		return pub_sub.ActionAccept
	}
	return pub_sub.ActionReject
}
