package handlers_test

import (
	"context"
	"duolingo/apps/message_input/server"
	"duolingo/apps/message_input/server/handlers/test/test_suites"
	"duolingo/dependencies"
	"duolingo/test/fixtures"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestMessageInputRequest(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.Bootstrap(context.Background(), "test", []string{
		"common",
		"connections",
		"message_queues",
	})

	server := server.NewMessageInputApiServer()

	suite.Run(t, test_suites.NewMessageInputRequestTestSuite(
		context.Background(),
		server,
	))
}
