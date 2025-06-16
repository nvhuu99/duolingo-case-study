package requests

import "errors"

type MessageInputRequest struct {
	campaign      string
	messageTitle  string
	messageBody   string
	validationErr error
}

func NewMessageInputRequest(
	campaign string,
	messageTitle string,
	messageBody string,
) (
	*MessageInputRequest,
	error,
) {
	req := &MessageInputRequest{
		campaign:     campaign,
		messageTitle: messageTitle,
		messageBody:  messageBody,
	}
	if !req.Validate() {
		return nil, req.GetValidationError()
	}
	return req, nil
}

func (req *MessageInputRequest) GetCampaign() string {
	return req.campaign
}

func (req *MessageInputRequest) GetMessageTitle() string {
	return req.messageTitle
}

func (req *MessageInputRequest) GetMessageBody() string {
	return req.messageBody
}

/* Implement interface Request */

func (req *MessageInputRequest) Validate() bool {
	if req.validationErr == nil {
		if req.campaign == "" {
			req.validationErr = errors.New("campaign must not empty")
		}
		if req.messageTitle == "" {
			req.validationErr = errors.New("message title must not empty")
		}
		if req.messageBody == "" {
			req.validationErr = errors.New("message body must not empty")
		}
	}
	return req.validationErr == nil
}

func (req *MessageInputRequest) GetValidationError() error {
	return req.validationErr
}
