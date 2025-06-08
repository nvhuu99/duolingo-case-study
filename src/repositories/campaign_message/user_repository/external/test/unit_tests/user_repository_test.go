package unit_tests

import (
	"context"
	"testing"
	"time"

	connection "duolingo/libraries/connection_manager/drivers/mongodb"
	mongodb_driver "duolingo/repositories/campaign_message/user_repository/drivers/mongodb"
	repo_test_suite "duolingo/repositories/campaign_message/user_repository/external/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserRepository(t *testing.T) {
	builder := connection.NewMongoConnectionBuilder(context.Background())
	builder.
		SetCredentials("root", "12345").
		SetHost("localhost").
		SetPort("27017").
		SetConnectionTimeOut(100 * time.Millisecond).
		SetConnectionRetryWait(10 * time.Millisecond).
		SetOperationReadTimeOut(100 * time.Millisecond).
		SetOperationWriteTimeOut(100 * time.Millisecond).
		SetOperationRetryWait(10 * time.Millisecond)
	_, err := builder.BuildConnectionManager()
	if err != nil {
		panic(err)
	}
	client, err := builder.BuildClientAndRegisterToManager()
	if err != nil {
		panic(err)
	}
	repo := mongodb_driver.NewUserRepo(client, "duolingo", "users")

	suite.Run(t, repo_test_suite.NewUserRepositoryTestSuite(repo))
}
