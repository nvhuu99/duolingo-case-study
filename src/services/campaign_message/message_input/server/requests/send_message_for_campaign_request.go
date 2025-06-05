package requests

import "errors"

type SendMessageForCampaignRequest struct {
	campaign      string
	messageTitle  string
	messageBody   string
	validationErr error
}

func NewSendMessageForCampaignRequest(
	campaign string,
	messageTitle string,
	messageBody string,
) *SendMessageForCampaignRequest {
	return &SendMessageForCampaignRequest{
		campaign:     campaign,
		messageTitle: messageTitle,
		messageBody:  messageBody,
	}
}

func (req *SendMessageForCampaignRequest) GetCampaign() string {
	return req.campaign
}

func (req *SendMessageForCampaignRequest) GetMessageTitle() string {
	return req.messageTitle
}

func (req *SendMessageForCampaignRequest) GetMessageBody() string {
	return req.messageBody
}

/* Implement interface Request */

func (req *SendMessageForCampaignRequest) Validate() bool {
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

func (req *SendMessageForCampaignRequest) GetValidationError() error {
	return req.validationErr
}
