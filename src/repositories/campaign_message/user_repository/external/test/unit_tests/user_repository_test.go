package unit_tests

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	mongodb_driver "duolingo/repositories/campaign_message/user_repository/drivers/mongodb"
// 	repo_test_suite "duolingo/repositories/campaign_message/user_repository/external/test/test_suites"

// 	"github.com/stretchr/testify/suite"
// )

// func TestMongoDBUserRepository(t *testing.T) {
// 	builder := mongodb_driver.NewUserRepoBuilder(context.Background()).
// 		SetCredentials("root", "12345").
// 		SetHost("localhost").
// 		SetPort("27017").
// 		SetConnectionTimeOut(5 * time.Second).
// 		SetOperationReadTimeOut(2 * time.Second).
// 		SetOperationWriteTimeOut(2 * time.Second).
// 		SetOperationRetryWait(10 * time.Millisecond).
// 		SetDatabaseName("duolingo").
// 		SetCollectionName("users")
// 	_, err := builder.BuildConnectionManager()
// 	if err != nil {
// 		panic(err)
// 	}
// 	client, err := builder.BuildClientAndRegisterToManager()
// 	if err != nil {
// 		panic(err)
// 	}
// 	repo, err := builder.BuildRepo(client)
// 	if err != nil {
// 		panic(err)
// 	}
// 	suite.Run(t, repo_test_suite.NewUserRepositoryTestSuite(repo))
// }
