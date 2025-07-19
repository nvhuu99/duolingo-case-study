package handlers_test

import (
	"context"
	"duolingo/apps/push_sender/server/test/test_suites"
	"duolingo/dependencies"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSender(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.RegisterDependencies(ctx)
	dependencies.BootstrapDependencies("push_sender")

	suite.Run(t, test_suites.NewSenderTestSuite())
}
