package mongodb

import (
	"context"
	"testing"

	mongo "duolingo/libraries/connection_manager/drivers/mongodb"
	facade "duolingo/libraries/connection_manager/facade"
	repo_driver "duolingo/repositories/user_repository/drivers/mongodb"
	"duolingo/repositories/user_repository/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserRepository(t *testing.T) {
	client := facade.Provider(context.Background()).InitMongo(mongo.
		DefaultMongoConnectionArgs().
		SetCredentials("root", "12345"),
	).GetMongoClient()

	repo := repo_driver.NewUserRepo(client, "duolingo", "users")

	suite.Run(t, test_suites.NewUserRepositoryTestSuite(repo))
}
