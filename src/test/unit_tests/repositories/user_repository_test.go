package mongodb

import (
	"context"
	"testing"

	"duolingo/dependencies"
	"duolingo/repositories/user_repository/external/test/test_suites"
	"duolingo/test/fixtures"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserRepository(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies("test", []string{
		"common",
		"connections",
		"user_repo",
		"user_service",
	})

	suite.Run(t, test_suites.NewUserRepositoryTestSuite())
}
