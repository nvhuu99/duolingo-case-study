package tasks

import (
	"context"
	"duolingo/constants"
	"duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/models"
	"duolingo/services/noti_builder/server/workloads"
)

func CreatePushNotiMessageBatches(ctx context.Context, job *workloads.TokenBatchJob) error {
	distributor := container.MustResolve[*workloads.TokenBatchDistributor]()
	publisher := container.MustResolve[pub_sub.Publisher]()
	return distributor.ConsumeIncomingBatches(ctx, job, func(
		input *models.MessageInput,
		devices []*models.UserDevice,
	) error {
		return publisher.Notify(constants.TopicPushNotiMessages,
			string(models.NewPushNotiMessage(input, devices).Encode()))
	})
}
