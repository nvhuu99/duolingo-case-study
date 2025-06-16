package test_suites

import (
	"duolingo/services/message_input/server/requests"

	"github.com/stretchr/testify/suite"
)

type MessageInputRequestTestSuite struct {
	suite suite.Suite
}

func (s *MessageInputRequestTestSuite) Test_Validation_Require_Params_Missing() {
	req1, _ := requests.NewMessageInputRequest("", "", "")
	s.suite.Assert().False(req1.Validate())
	s.suite.Assert().Error(req1.GetValidationError())

	req2, _ := requests.NewMessageInputRequest("campaign_name", "", "")
	s.suite.Assert().False(req2.Validate())
	s.suite.Assert().Error(req2.GetValidationError())

	req3, _ := requests.NewMessageInputRequest("campaign_name", "message_title", "")
	s.suite.Assert().False(req3.Validate())
	s.suite.Assert().Error(req3.GetValidationError())

	req4, _ := requests.NewMessageInputRequest("campaign_name", "message_title", "message_body")
	s.suite.Assert().True(req4.Validate())
	s.suite.Assert().NoError(req4.GetValidationError())
}

func (s *MessageInputRequestTestSuite) Test_Create_Request() {
	req, validationErr := requests.NewMessageInputRequest(
		"campaign_name",
		"message_title",
		"message_body",
	)
	s.suite.Assert().NoError(validationErr)
	s.suite.Assert().Equal("campaign_name", req.GetCampaign())
	s.suite.Assert().Equal("message_title", req.GetMessageTitle())
	s.suite.Assert().Equal("message_body", req.GetMessageBody())
}
