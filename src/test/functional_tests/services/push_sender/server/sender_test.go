package handlers_test

import (
	"context"
	"duolingo/apps/push_sender/server/test/test_suites"
	"duolingo/dependencies"
	"duolingo/test/fixtures"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSender(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies("test", []string{
		"common",
		"connections",
		"message_queues",
		"push_service",
	})

	suite.Run(t, test_suites.NewSenderTestSuite())
}
