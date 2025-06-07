package test_suites

import (
	"context"

	"duolingo/libraries/mongo_connect"

	"github.com/stretchr/testify/suite"
)

type ConnectionBuilderTestSuite struct {
	suite.Suite
	builder *mongo_connect.ConnectionBuilder
}

func (s *ConnectionBuilderTestSuite) SetupTest() {
	s.builder = mongo_connect.NewConnectionBuilder(context.Background())
}

func (s *ConnectionBuilderTestSuite) TearDownTest() {
	s.builder.Destroy()
	s.builder = nil
}

func (s *ConnectionBuilderTestSuite) TestEnforceConnectionManagerSingleton() {
	_, firstBuildErr := s.builder.BuildConnectionManager()
	_, secondBuildErr := s.builder.BuildConnectionManager()
	s.Assert().NoError(firstBuildErr)
	s.Assert().Equal(mongo_connect.ErrConnManagerSingletonViolation, secondBuildErr)
}

func (s *ConnectionBuilderTestSuite) TestBuildWithoutConnectionManagerErr() {
	_, buildClientErr := s.builder.BuildClientAndRegisterToManager()
	s.Assert().Equal(buildClientErr, mongo_connect.ErrConnManagerHasNotCreated)
}

func (s *ConnectionBuilderTestSuite) TestBuildConnectionManagerAndClient() {
	manager, buildManagerErr := s.builder.BuildConnectionManager()
	if !s.Assert().NotNil(manager) || !s.Assert().NoError(buildManagerErr) {
		return
	}
	client, buildClientErr := s.builder.BuildClientAndRegisterToManager()
	s.Assert().NotNil(client)
	s.Assert().NoError(buildClientErr)
}
