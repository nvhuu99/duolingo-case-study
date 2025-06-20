package test_suites

import (
	"context"
	driver "duolingo/libraries/push_notification/drivers/firebase"
	driver_msg "duolingo/libraries/push_notification/drivers/firebase/message"

	"github.com/stretchr/testify/suite"
)

type PushNotiFactoryTestSuite struct {
	suite.Suite
	credJson string
}

func NewPushNotiFactoryTestSuite(credJson string) *PushNotiFactoryTestSuite {
	return &PushNotiFactoryTestSuite{
		credJson: credJson,
	}
}

func (s *PushNotiFactoryTestSuite) Test_CreateFactory() {
	factory, err := driver.NewFirebasePushNotiFactory(context.TODO(), s.credJson)
	s.Assert().NotNil(factory)
	s.Assert().NoError(err)
}

func (s *PushNotiFactoryTestSuite) Test_CreateMessageBuilder() {
	factory, _ := driver.NewFirebasePushNotiFactory(context.TODO(), s.credJson)
	builder := factory.CreateMessageBuilder()
	_, isFirebaseFamily := builder.(*driver_msg.FirebaseMessagebuilder)

	s.Assert().NotNil(builder)
	s.Assert().True(isFirebaseFamily)
}

func (s *PushNotiFactoryTestSuite) Test_CreatePushService() {
	factory, _ := driver.NewFirebasePushNotiFactory(context.TODO(), s.credJson)
	service, err := factory.CreatePushService()
	_, isFirebaseFamily := service.(*driver.FirebasePushService)

	s.Assert().NoError(err)
	s.Assert().NotNil(service)
	s.Assert().True(isFirebaseFamily)
}
