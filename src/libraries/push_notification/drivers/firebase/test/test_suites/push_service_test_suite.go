package test_suites

import (
	"context"
	"duolingo/libraries/push_notification/drivers/firebase"
	"duolingo/libraries/push_notification/message"

	"github.com/stretchr/testify/suite"
)

type PushServiceTestSuite struct {
	suite.Suite
	service *firebase.FirebasePushService
}

func NewPushServiceTestSuite(service *firebase.FirebasePushService) *PushServiceTestSuite {
	return &PushServiceTestSuite{service: service}
}

func (s *PushServiceTestSuite) Test_SendMulticast_For_Android() {
	msg := &message.Message{
		Title: "test",
		Body:  "test",
	}
	target := &message.MulticastTarget{
		DeviceTokens: []string{"fakeToken1", "fakeToken2", "fakeToken3"},
		Platforms:    []message.Platform{message.Android, message.IOS},
	}
	result, err := s.service.SendMulticast(context.Background(), msg, target)

	s.Assert().NoError(err)
	s.Assert().Equal(len(target.DeviceTokens), result.FailureCount+result.SuccessCount)
}
