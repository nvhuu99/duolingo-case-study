package handlers_test

import (
	"duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/services/message_input/server/handlers/test/test_suites"
	"duolingo/test/fixtures/bootstrap"
	cnst "duolingo/test/fixtures/constants"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestRequestHandlers(t *testing.T) {
	bootstrap.Bootstrap()

	inputPublisher := container.MustResolveAlias[pub_sub.Publisher](cnst.MesgInputPublisher)
	inputSubscriber := container.MustResolveAlias[pub_sub.Subscriber](cnst.MesgInputSubscriber)

	suite.Run(t, test_suites.NewHandlersTestSuite(inputPublisher, inputSubscriber))
}
