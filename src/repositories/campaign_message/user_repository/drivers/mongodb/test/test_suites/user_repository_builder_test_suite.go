package test_suites

import (
	"context"
	user_repo "duolingo/repositories/campaign_message/user_repository/drivers/mongodb"

	"github.com/stretchr/testify/suite"
)

type UserRepoBuilderTestSuite struct {
	suite.Suite
}

func (s *UserRepoBuilderTestSuite) TestEnforceConnectionManagerSingleton() {
	builder := user_repo.NewUserRepoBuilder(context.Background())
	_, firstBuildErr := builder.BuildConnectionManager()
	_, secondBuildErr := builder.BuildConnectionManager()
	s.Assert().NoError(firstBuildErr)
	s.Assert().Equal(user_repo.ErrConnManagerSingletonViolation, secondBuildErr)
}

func (s *UserRepoBuilderTestSuite) TestBuildWithoutConnectionManagerErr() {
	builder := user_repo.NewUserRepoBuilder(context.Background())
	_, buildClientErr := builder.BuildClientAndRegisterToManager()
	_, buildRepoErr := builder.BuildRepo(nil)
	s.Assert().Equal(buildClientErr, user_repo.ErrConnManagerHasNotCreated)
	s.Assert().Equal(buildRepoErr, user_repo.ErrConnManagerHasNotCreated)
}

func (s *UserRepoBuilderTestSuite) TestEnforceUserRepoSingleton() {
	builder := user_repo.NewUserRepoBuilder(context.Background())
	builder.BuildConnectionManager()
	client, _ := builder.BuildClientAndRegisterToManager()
	_, firstBuildErr := builder.BuildRepo(client)
	_, secondBuildErr := builder.BuildRepo(client)
	s.Assert().NoError(firstBuildErr)
	s.Assert().Equal(secondBuildErr, user_repo.ErrUserRepoSingletonViolation)
}
