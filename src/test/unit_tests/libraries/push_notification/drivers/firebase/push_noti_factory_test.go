package firebase

import (
	"context"
	"testing"

	"duolingo/dependencies"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	"duolingo/libraries/push_notification/drivers/firebase/test/test_suites"
	"duolingo/test/fixtures"

	"github.com/stretchr/testify/suite"
)

func TestPushNotiFactory(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies("test", []string{
		"common",
	})

	config := container.MustResolve[config_reader.ConfigReader]()
	cred := config.Get("firebase", "credentials")

	suite.Run(t, test_suites.NewPushNotiFactoryTestSuite(cred))
}
