package handlers

import (
	"errors"

	"duolingo/constants"
	pubsub "duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	models "duolingo/models"
	"duolingo/services/message_input/server/requests"
)

func HandleMessageInputRequest(req *requests.MessageInputRequest) error {
	publisher := container.MustResolve[pubsub.Publisher]()
	message := models.NewMessageInput(
		req.GetCampaign(),
		req.GetMessageTitle(),
		req.GetMessageBody(),
	)
	err := publisher.Notify(
		constants.TopicMessageInputs,
		string(message.Encode()),
	)
	if err != nil {
		return errors.New("failed to notify message input request")
	}
	return nil
}
