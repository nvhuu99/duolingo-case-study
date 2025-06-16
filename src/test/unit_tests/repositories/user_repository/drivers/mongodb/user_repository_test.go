package mongodb

import (
	"context"
	"testing"

	connection "duolingo/libraries/connection_manager/drivers/mongodb"
	facade "duolingo/libraries/connection_manager/facade"
	mongodb "duolingo/repositories/user_repository/drivers/mongodb"
	"duolingo/repositories/user_repository/external/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserRepository(t *testing.T) {
	client := facade.Provider(context.Background()).InitMongo(connection.
		DefaultMongoConnectionArgs().
		SetCredentials("root", "12345"),
	).GetMongoClient()

	factory := mongodb.NewUserRepoFactory(client)

	suite.Run(t, test_suites.NewUserRepositoryTestSuite(factory))
}
