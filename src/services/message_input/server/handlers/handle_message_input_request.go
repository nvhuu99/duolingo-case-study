package handlers

import (
	"errors"

	"duolingo/constants"
	pubsub "duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	dist "duolingo/libraries/work_distributor"
	models "duolingo/models"
	usr_repo "duolingo/repositories/user_repository"
	"duolingo/services/message_input/server/requests"
)

func HandleMessageInputRequest(req *requests.MessageInputRequest) error {
	var err error
	var devicesCount uint64
	var workload *dist.Workload
	if devicesCount, err = countUserDevicesForCampaign(req.GetCampaign()); err == nil {
		if workload, err = createWorkloadForUserDevicesBatching(devicesCount); err == nil {
			err = notifyOfNewMessageAsTheWorkloadInitilization(req, workload)
		}
	}
	return err
}

func countUserDevicesForCampaign(campaign string) (uint64, error) {
	userRepo := container.MustResolve[usr_repo.UserRepository]()
	devicesCount, err := userRepo.CountUserDevicesForCampaign(campaign)
	if err != nil {
		return 0, errors.New("failed to get devices count for campaign")
	}
	return devicesCount, nil
}

func createWorkloadForUserDevicesBatching(userDeviceCount uint64) (
	*dist.Workload,
	error,
) {
	distributor := container.MustResolve[*dist.WorkDistributor]()
	workload, err := distributor.CreateWorkload(userDeviceCount)
	if err != nil {
		return nil, errors.New("failed to create workload")
	}
	return workload, nil
}

func notifyOfNewMessageAsTheWorkloadInitilization(
	req *requests.MessageInputRequest,
	workload *dist.Workload,
) error {
	publisher := container.MustResolve[pubsub.Publisher]()
	message := models.NewInputMessage(
		req.GetCampaign(),
		req.GetMessageTitle(),
		req.GetMessageBody(),
		workload.Id,
	)
	err := publisher.Notify(
		constants.PubSubTopicMessageInput, 
		string(message.Encode()),
	)
	if err != nil {
		return errors.New("failed to notify message request")
	}
	return nil
}
