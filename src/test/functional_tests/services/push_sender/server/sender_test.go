package handlers_test

import (
	ps "duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/services/push_sender/server/test/test_suites"
	"duolingo/test/fixtures/bootstrap"
	cnst "duolingo/test/fixtures/constants"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSender(t *testing.T) {
	bootstrap.Bootstrap()

	notiPublisher := container.MustResolveAlias[ps.Publisher](cnst.PushNotiPublisher)
	notiSubscriber := container.MustResolveAlias[ps.Subscriber](cnst.PushNotiSubscriber)

	suite.Run(t, test_suites.NewSenderTestSuite(notiPublisher, notiSubscriber))
}
