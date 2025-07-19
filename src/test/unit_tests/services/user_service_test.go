package mongodb

import (
	"context"
	"testing"

	"duolingo/dependencies"
	"duolingo/services/user_service/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.RegisterDependencies(ctx)
	dependencies.BootstrapDependencies()

	suite.Run(t, test_suites.NewUserServiceTestSuite())
}
