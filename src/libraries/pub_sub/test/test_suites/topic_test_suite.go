package test_suites

import (
	"duolingo/libraries/pub_sub"

	"github.com/stretchr/testify/suite"
)

type TopicTestSuite struct {
	suite.Suite
}

func (s *PublisherTestSuite) TestEqualTopic() {
	nilCompare := (pub_sub.NewTopic("")).Equal(nil)
	diffCompare := (pub_sub.NewTopic("abc")).Equal(pub_sub.NewTopic("def"))
	sameCompare := (pub_sub.NewTopic("abc")).Equal(pub_sub.NewTopic("abc"))
	s.Assert().False(nilCompare)
	s.Assert().False(diffCompare)
	s.Assert().True(sameCompare)
}
