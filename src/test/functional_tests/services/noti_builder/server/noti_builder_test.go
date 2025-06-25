package handlers_test

import (
	ps "duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/repositories/user_repository/external"
	"duolingo/services/noti_builder/server"
	"duolingo/services/noti_builder/server/test/test_suites"
	"duolingo/services/noti_builder/server/workloads"
	"duolingo/test/fixtures/bootstrap"
	cnst "duolingo/test/fixtures/constants"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestNotiBuilder(t *testing.T) {
	bootstrap.Bootstrap()

	userRepo := container.MustResolveAlias[external.UserRepository](cnst.UserRepo)
	inputPublisher := container.MustResolveAlias[ps.Publisher](cnst.MesgInputPublisher)
	inputSubscriber := container.MustResolveAlias[ps.Subscriber](cnst.MesgInputSubscriber)
	notiPublisher := container.MustResolveAlias[ps.Publisher](cnst.PushNotiPublisher)
	notiSubscriber := container.MustResolveAlias[ps.Subscriber](cnst.PushNotiSubscriber)
	tokenDistributor := container.MustResolve[*workloads.TokenBatchDistributor]()
	builder := server.NewNotiBuilder(inputSubscriber, notiPublisher, tokenDistributor)

	suite.Run(t, test_suites.NewNotiBuilderTestSuite(
		userRepo,
		inputPublisher,
		notiSubscriber,
		builder,
	))
}
