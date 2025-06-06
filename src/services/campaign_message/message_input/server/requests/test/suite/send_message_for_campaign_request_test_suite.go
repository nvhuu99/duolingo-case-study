package suite

import (
	"duolingo/services/campaign_message/message_input/server/requests"

	"github.com/stretchr/testify/suite"
)

type SendMessageForCampaignRequestTestSuite struct {
	suite suite.Suite
}

func (s *SendMessageForCampaignRequestTestSuite) TestCreateRequest() {
	req := requests.NewSendMessageForCampaignRequest(
		"campaign_name",
		"message_title",
		"message_body",
	)
	s.suite.Assert().Equal("campaign_name", req.GetCampaign())
	s.suite.Assert().Equal("message_title", req.GetMessageTitle())
	s.suite.Assert().Equal("message_body", req.GetMessageBody())
}

func (s *SendMessageForCampaignRequestTestSuite) TestRequestValidations() {
	req1 := requests.NewSendMessageForCampaignRequest("", "", "")
	s.suite.Assert().False(req1.Validate())
	s.suite.Assert().Error(req1.GetValidationError())

	req2 := requests.NewSendMessageForCampaignRequest("campaign_name", "", "")
	s.suite.Assert().False(req2.Validate())
	s.suite.Assert().Error(req2.GetValidationError())

	req3 := requests.NewSendMessageForCampaignRequest("campaign_name", "message_title", "")
	s.suite.Assert().False(req3.Validate())
	s.suite.Assert().Error(req3.GetValidationError())

	req4 := requests.NewSendMessageForCampaignRequest("campaign_name", "message_title", "message_body")
	s.suite.Assert().True(req4.Validate())
	s.suite.Assert().NoError(req4.GetValidationError())
}
