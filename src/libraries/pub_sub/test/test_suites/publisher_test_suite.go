package test_suites

import (
	"duolingo/libraries/pub_sub"

	"github.com/stretchr/testify/suite"
)

type PublisherTestSuite struct {
	suite.Suite
	publisher pub_sub.Publisher
}

func (s *PublisherTestSuite) TestNotifyNilTopic() {
	err := s.publisher.Notify(nil, struct{}{})
	s.Assert().Error(err)
}

func (s *PublisherTestSuite) TestNotifyEmptyTopic() {
	err := s.publisher.Notify(pub_sub.NewTopic(""), struct{}{})
	s.Assert().NoError(err)
}

func (s *PublisherTestSuite) TestNotifyTopic() {
	err := s.publisher.Notify(pub_sub.NewTopic("test"), struct{}{})
	s.Assert().NoError(err)
}
