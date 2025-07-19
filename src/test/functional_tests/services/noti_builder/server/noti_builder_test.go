package handlers_test

import (
	"context"
	"duolingo/apps/noti_builder/server"
	"duolingo/apps/noti_builder/server/test/test_suites"
	"duolingo/dependencies"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestNotiBuilder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.RegisterDependencies(ctx)
	dependencies.BootstrapDependencies("noti_builder")

	suite.Run(t, test_suites.NewNotiBuilderTestSuite(server.NewNotiBuilder()))
}
