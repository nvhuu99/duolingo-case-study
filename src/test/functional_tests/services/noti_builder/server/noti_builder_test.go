package handlers_test

import (
	"context"
	"duolingo/apps/noti_builder/server/test/test_suites"
	"duolingo/dependencies"
	"duolingo/test/fixtures"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestNotiBuilder(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.Bootstrap(context.Background(), "test", []string{
		"common",
		"connections",
		"message_queues",
		"user_repo",
		"user_service",
		"work_distributor",
	})
	suite.Run(t, test_suites.NewNotiBuilderTestSuite())
}
