package test_suites

import (
	"context"
	"duolingo/libraries/connection_manager"
	"duolingo/libraries/connection_manager/test/fake"

	"github.com/stretchr/testify/suite"
)

type ConnectionBuilderTestSuite struct {
	suite.Suite
	builder *connection_manager.ConnectionBuilder
}

func (s *ConnectionBuilderTestSuite) SetupTest() {
	s.builder = connection_manager.NewConnectionBuilder(context.Background())
	s.builder.SetConnectionDriver(fake.NewFakeConnectionProxy())
}

func (s *ConnectionBuilderTestSuite) TearDownTest() {
	s.builder.Destroy()
	s.builder = nil
}

func (s *ConnectionBuilderTestSuite) TestEnforceConnectionManagerSingleton() {
	_, firstBuildErr := s.builder.BuildConnectionManager()
	_, secondBuildErr := s.builder.BuildConnectionManager()
	s.Assert().NoError(firstBuildErr)
	s.Assert().Equal(connection_manager.ErrConnManagerSingletonViolation, secondBuildErr)
}

func (s *ConnectionBuilderTestSuite) TestBuildWithoutConnectionManagerErr() {
	_, buildClientErr := s.builder.BuildClientAndRegisterToManager()
	s.Assert().Equal(buildClientErr, connection_manager.ErrConnManagerHasNotCreated)
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
