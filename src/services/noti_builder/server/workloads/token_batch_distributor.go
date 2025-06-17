package workloads

import (
	"context"
	"duolingo/constants"
	"duolingo/libraries/pub_sub"
	"duolingo/libraries/work_distributor"
	"duolingo/models"
	"duolingo/repositories/user_repository/external/services"
	"time"
)

type TokenBatchDistributor struct {
	*work_distributor.WorkDistributor
	services.UserService
	pub_sub.Publisher
}

func NewTokenBatchDistributor(
	base *work_distributor.WorkDistributor,
	userService services.UserService,
	publisher pub_sub.Publisher,
) *TokenBatchDistributor {
	return &TokenBatchDistributor{
		WorkDistributor: base,
		UserService:     userService,
		Publisher:       publisher,
	}
}

func (d *TokenBatchDistributor) CreateBatchJob(input *models.MessageInput) error {
	var err error
	var count uint64
	var workload *work_distributor.Workload
	if count, err = d.CountDevicesForCampaign(input.Campaign); err == nil && count != 0 {
		if workload, err = d.CreateWorkload(count); err == nil {
			return d.Notify(constants.TopicNotiBuilderJobs,
				string(NewTokenBatchJob(workload.Id, input).Encode()))
		}
	}
	return err
}

func (d *TokenBatchDistributor) ConsumeIncomingBatches(
	ctx context.Context,
	job *TokenBatchJob,
	closure func(input *models.MessageInput, deviceTokens []string) error,
) error {
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
			lastErr = fulfilledOrErr(lastErr)
			continue
		}
		lastErr = d.HandleAssignment(assignment, func() error {
			tokens, tokenErr := d.GetDeviceTokensForCampaign(
				job.Message.Campaign,
				assignment.WorkStartAt(),
				assignment.WorkEndAt(),
			)
			if tokenErr != nil {
				return tokenErr
			}
			return closure(job.Message, tokens)
		})
	}
}

func fulfilledOrErr(err error) error {
	if err == work_distributor.ErrWorkloadHasAlreadyFulfilled {
		return nil
	}
	return err
}
