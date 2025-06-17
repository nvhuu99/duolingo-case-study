package tasks

import (
	container "duolingo/libraries/service_container"
	"duolingo/models"
	"duolingo/services/noti_builder/server/workloads"
)

func HandleMessageInput(input *models.MessageInput) error {
	distributor := container.MustResolve[*workloads.TokenBatchDistributor]()
	return distributor.CreateBatchJob(input)
}
