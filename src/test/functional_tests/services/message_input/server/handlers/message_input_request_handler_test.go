package handlers_test

import (
	"context"
	"duolingo/apps/message_input/server"
	"duolingo/apps/message_input/server/handlers/test/test_suites"
	"duolingo/dependencies"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestMessageInputRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.RegisterDependencies(ctx)
	dependencies.BootstrapDependencies("message_input")

	server := server.NewMessageInputApiServer()
	server.RegisterRoutes()

	suite.Run(t, test_suites.NewMessageInputRequestTestSuite(ctx, server))
}
