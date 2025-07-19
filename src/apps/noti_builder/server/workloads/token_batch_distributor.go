package workloads

import (
	"context"
	"time"

	container "duolingo/libraries/dependencies_container"
	ps "duolingo/libraries/message_queue/pub_sub"
	dist "duolingo/libraries/work_distributor"
	"duolingo/models"
	usr_svc "duolingo/services/user_service"
)

type TokenBatchDistributor struct {
	*dist.WorkDistributor

	jobPublisher  ps.Publisher
	jobSubscriber ps.Subscriber

	userService usr_svc.UserService
}

func NewTokenBatchDistributor() *TokenBatchDistributor {
	return &TokenBatchDistributor{
		WorkDistributor: container.MustResolve[*dist.WorkDistributor](),
		jobPublisher:    container.MustResolveAlias[ps.Publisher]("noti_builder_jobs_publisher"),
		jobSubscriber:   container.MustResolveAlias[ps.Subscriber]("noti_builder_jobs_subscriber"),
		userService:     container.MustResolve[usr_svc.UserService](),
	}
}

func (d *TokenBatchDistributor) CreateBatchJob(input *models.MessageInput) error {
	var err error
	var count int64
	var workload *dist.Workload
	if count, err = d.userService.CountDevicesForCampaign(input.Campaign); count != 0 && err == nil {
		if workload, err = d.CreateWorkload(count); err == nil {
			job := NewTokenBatchJob(workload.Id, input)
			err = d.jobPublisher.NotifyMainTopic(string(job.Encode()))
		}
	}
	return err
}

func (d *TokenBatchDistributor) ConsumingTokenBatches(
	ctx context.Context,
	batchConsumer func(input *models.MessageInput, devices []*models.UserDevice),
) error {
	return d.jobSubscriber.ListeningMainTopic(ctx, func(ctx context.Context, str string) {
		d.startJobBatching(ctx, JobDecode([]byte(str)), batchConsumer)
	})
}

func (d *TokenBatchDistributor) startJobBatching(
	ctx context.Context,
	job *TokenBatchJob,
	closure func(input *models.MessageInput, devices []*models.UserDevice),
) error {
	if jobErr := job.Validate(); jobErr != nil {
		return jobErr
	}

	consumeCtx, consumeCancel := context.WithCancel(ctx)
	defer consumeCancel()

	var lastErr error
	var interval = 10 * time.Millisecond
	var jobId = job.JobId
	var assignment *dist.Assignment
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
			if lastErr == dist.ErrWorkloadHasAlreadyFulfilled {
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
			closure(job.Message, devices)
			return nil
		})
	}
}
