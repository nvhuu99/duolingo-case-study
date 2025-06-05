package handlers

import (
	"errors"

	constants "duolingo/constants/campaign_message"
	"duolingo/libraries/pub_sub"
	"duolingo/libraries/work_distributor"
	models "duolingo/models/campaign_message"
	"duolingo/services/campaign_message/message_input/bootstrap"
	"duolingo/services/campaign_message/message_input/server/requests"
)

func HandleSendMessageForCampaignRequest(req *requests.SendMessageForCampaignRequest) error {
	if !req.Validate() {
		return req.GetValidationError()
	}
	var err error
	var devicesCount uint64
	var workload *work_distributor.Workload
	if devicesCount, err = countUserDevicesForCampaign(req.GetCampaign()); err == nil {
		if workload, err = createWorkloadForUserDevicesBatching(devicesCount); err == nil {
			err = notifyOfNewMessageAsTheWorkloadInitilization(req, workload)
		}
	}
	return err
}

func countUserDevicesForCampaign(campaign string) (uint64, error) {
	userRepo := bootstrap.GetUserRepository()
	if userRepo == nil {
		return 0, errors.New("failed to resolve user repository")
	}
	devicesCount, err := userRepo.CountUserDevicesForCampaign(campaign)
	if err != nil {
		return 0, errors.New("failed to get devices count for campaign")
	}
	return devicesCount, nil
}

func createWorkloadForUserDevicesBatching(userDeviceCount uint64) (
	*work_distributor.Workload,
	error,
) {
	distributor := bootstrap.GetWorkDistributor()
	if distributor == nil {
		return nil, errors.New("failed to resolve work distributor")
	}
	workload, err := distributor.CreateWorkload(userDeviceCount)
	if err != nil {
		return nil, errors.New("failed to create workload")
	}
	return workload, nil
}

func notifyOfNewMessageAsTheWorkloadInitilization(
	req *requests.SendMessageForCampaignRequest,
	workload *work_distributor.Workload,
) error {
	publisher := bootstrap.GetPublisher()
	if publisher == nil {
		return errors.New("failed to resolve publisher")
	}
	message := models.NewInputMessage(
		req.GetCampaign(),
		req.GetMessageTitle(),
		req.GetMessageBody(),
		workload.GetId(),
	)
	err := publisher.Notify(
		pub_sub.NewTopic(constants.PUBSUB_TOPIC_CAMPAIGN_INPUT_MESSAGE),
		message,
	)
	if err != nil {
		return errors.New("failed to notify message request")
	}
	return nil
}
