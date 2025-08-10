package workloads

import (
	"context"
	"log"
	"time"

	container "duolingo/libraries/dependencies_container"
	events "duolingo/libraries/events/facade"
	ps "duolingo/libraries/message_queue/pub_sub"
	dist "duolingo/libraries/work_distributor"
	"duolingo/models"
	usr_svc "duolingo/services/user_service"
)

type TokenBatchDistributor struct {
	*dist.WorkDistributor

	buildJobPublisher  ps.Publisher
	buildJobSubscriber ps.Subscriber

	userService *usr_svc.UserService
}

func NewTokenBatchDistributor() *TokenBatchDistributor {
	return &TokenBatchDistributor{
		WorkDistributor:    container.MustResolve[*dist.WorkDistributor](),
		buildJobPublisher:  container.MustResolveAlias[ps.Publisher]("noti_builder_jobs_publisher"),
		buildJobSubscriber: container.MustResolveAlias[ps.Subscriber]("noti_builder_jobs_subscriber"),
		userService:        container.MustResolve[*usr_svc.UserService](),
	}
}

func (d *TokenBatchDistributor) CreateBatchJob(ctx context.Context, input *models.MessageInput) error {
	var err error
	var count int64
	var workload *dist.Workload

	evt := events.Start(ctx, "token_batch_distributor.create_batch_job", nil)
	defer events.End(evt, true, err, nil)
	defer log.Println("token_distributor:create batch job, err:", err)

	if count, err = d.userService.CountDevicesForCampaign(evt.Context(), input.Campaign); err == nil {
		if count == 0 {
			log.Println("token_distributor: workload empty")
			return nil
		}
		if workload, err = d.CreateWorkload(evt.Context(), count); err == nil {
			job := NewTokenBatchJob(workload.Id, input)
			err = d.buildJobPublisher.NotifyMainTopic(evt.Context(), string(job.Encode()))
			evt.SetData("devices_total", workload.TotalWorkUnits)
			evt.SetData("batch_size", workload.TotalUnitsPerAssignment)
			evt.SetData("expected_batches_total", workload.GetExpectTotalAssignments())
		}
	}
	return err
}

func (d *TokenBatchDistributor) ConsumingTokenBatches(
	ctx context.Context,
	batchConsumer func(
		ctx context.Context,
		input *models.MessageInput,
		devices []*models.UserDevice,
	) error,
) error {
	return d.buildJobSubscriber.ListeningMainTopic(ctx, func(ctx context.Context, str string) error {
		return d.startJobBatching(ctx, JobDecode([]byte(str)), batchConsumer)
	})
}

func (d *TokenBatchDistributor) startJobBatching(
	ctx context.Context,
	job *TokenBatchJob,
	batchReceiver func(
		ctx context.Context,
		input *models.MessageInput,
		devices []*models.UserDevice,
	) error,
) error {
	evt := events.Start(ctx, "token_batch_distributor.job_batching", nil)
	defer log.Println("start job batching")

	if jobErr := job.Validate(); jobErr != nil {
		events.Failed(evt, jobErr, nil)
		return jobErr
	}

	consumeCtx, consumeCancel := context.WithCancel(ctx)
	defer consumeCancel()

	var lastErr error
	var interval = 10 * time.Millisecond
	var jobId = job.JobId
	var assignment *dist.Assignment
	var batchCount int
	for {
		select {
		case <-consumeCtx.Done():
			return nil
		default:
		}
		if lastErr != nil {
			events.Failed(evt, lastErr, nil)
			return lastErr
		}
		if assignment, lastErr = d.WaitForAssignment(consumeCtx, interval, jobId); lastErr != nil {
			if lastErr == dist.ErrWorkloadHasAlreadyFulfilled {
				events.Succeeded(evt, nil)
				return nil
			}
			continue
		}

		batchCount++
		evt.SetData("batch_count", batchCount)

		lastErr = d.HandleAssignment(evt.Context(), assignment, func(assignmentCtx context.Context) error {
			devices, queryErr := d.userService.GetDevicesForCampaign(
				assignmentCtx,
				job.Message.Campaign,
				assignment.WorkStartAt()-1, // offset
				assignment.WorkEndAt()-assignment.WorkStartAt()+1, // limit
			)
			if queryErr != nil {
				return queryErr
			}
			return batchReceiver(assignmentCtx, job.Message, devices)
		})
	}
}
