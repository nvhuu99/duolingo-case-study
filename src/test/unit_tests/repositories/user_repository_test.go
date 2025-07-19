package mongodb

import (
	"context"
	"testing"

	"duolingo/dependencies"
	"duolingo/repositories/user_repository/external/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserRepository(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.RegisterDependencies(ctx)
	dependencies.BootstrapDependencies()

	suite.Run(t, test_suites.NewUserRepositoryTestSuite())
}
