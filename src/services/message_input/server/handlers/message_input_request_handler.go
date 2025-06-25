package handlers

import (
	"errors"

	"duolingo/libraries/pub_sub"
	"duolingo/models"
	"duolingo/services/message_input/server/requests"
)

type MessageInputRequestHandler struct {
	inputPublisher pub_sub.Publisher
}

func NewMessageInputRequestHandler(inputPublisher pub_sub.Publisher) *MessageInputRequestHandler {
	return &MessageInputRequestHandler{inputPublisher}
}

func (handler *MessageInputRequestHandler) Handle(req *requests.MessageInputRequest) (
	*models.MessageInput,
	error,
) {
	if !req.Validate() {
		return nil, req.GetValidationError()
	}
	message := models.NewMessageInput(
		req.GetCampaign(),
		req.GetMessageTitle(),
		req.GetMessageBody(),
	)
	err := handler.inputPublisher.NotifyMainTopic(string(message.Encode()))
	if err != nil {
		return nil, errors.New("failed to notify message input request")
	}
	return message, nil
}
