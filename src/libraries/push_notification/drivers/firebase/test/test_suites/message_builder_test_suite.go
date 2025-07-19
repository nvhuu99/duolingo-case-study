package test_suites

import (
	driver "duolingo/libraries/push_notification/drivers/firebase/message"
	"duolingo/libraries/push_notification/message"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/suite"
)

type MessageBuilderTestSuite struct {
	suite.Suite
}

func (s *MessageBuilderTestSuite) Test_Message_Validate() {
	msg := message.Message{}
	validation := msg.Validate()
	s.Assert().Equal(validation, message.ErrMessageRequireParamsMissing)
}

func (s *MessageBuilderTestSuite) Test_MulticastTarget_Validate() {
	target := message.MulticastTarget{}
	validation := target.Validate()
	s.Assert().Equal(validation, message.ErrMulticastTargetInadequate)
}

func (s *MessageBuilderTestSuite) Test_BuildMulticast_For_Android() {
	msg := &message.Message{
		Title:       "title",
		Body:        "content",
		Icon:        "icon",
		Sound:       "sound",
		Expiration:  10 * time.Minute,
		CollapseKey: "collapseKey",
		Priority:    message.PriorityHigh,
		Visibility:  message.VisibilityPrivate,
	}
	target := &message.MulticastTarget{
		DeviceTokens: []string{"fake_token"},
		Platforms:    []message.Platform{message.Android, message.IOS},
	}
	multicast, err := driver.NewFirebaseMessagebuilder().BuildMulticast(msg, target)

	firebaseMulticast, ok := multicast.(*messaging.MulticastMessage)

	s.Assert().NoError(err)
	s.Assert().NotNil(firebaseMulticast)
	s.Assert().True(ok)

	s.Assert().Equal("title", firebaseMulticast.Notification.Title)
	s.Assert().Equal("content", firebaseMulticast.Notification.Body)
	s.Assert().Equal("icon", firebaseMulticast.Android.Notification.Icon)
	s.Assert().Equal("sound", firebaseMulticast.Android.Notification.Sound)
	s.Assert().Equal(10*time.Minute, *firebaseMulticast.Android.TTL)
	s.Assert().Equal("collapseKey", firebaseMulticast.Android.CollapseKey)
	s.Assert().Equal("high", firebaseMulticast.Android.Priority)
	s.Assert().Equal(messaging.VisibilityPrivate, firebaseMulticast.Android.Notification.Visibility)
}
