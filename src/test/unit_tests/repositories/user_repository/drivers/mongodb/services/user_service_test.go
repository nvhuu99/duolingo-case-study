package mongodb

import (
	"testing"

	container "duolingo/libraries/service_container"
	"duolingo/repositories/user_repository/drivers/mongodb"
	"duolingo/repositories/user_repository/drivers/mongodb/services"
	"duolingo/repositories/user_repository/external/test/test_suites"
	"duolingo/test/fixtures/bootstrap"
	cnst "duolingo/test/fixtures/constants"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserService(t *testing.T) {
	bootstrap.Bootstrap()

	factory := container.MustResolveAlias[*mongodb.UserRepoFactory](cnst.UserRepoFactory)
	repo := container.MustResolveAlias[*mongodb.UserRepo](cnst.UserRepo)
	service := container.MustResolveAlias[*services.UserService](cnst.UserService)

	suite.Run(t, test_suites.NewUserServiceTestSuite(
		factory,
		repo,
		service,
	))
}
