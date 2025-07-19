package mongodb

import (
	"context"
	"testing"

	"duolingo/dependencies"
	"duolingo/services/user_service/test/test_suites"
	"duolingo/test/fixtures"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserService(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies(
		"common",
		"connections",
		"user_repo",
		"user_service",
	)

	suite.Run(t, test_suites.NewUserServiceTestSuite())
}
