package redis

import (
	"context"
	"testing"

	"duolingo/dependencies"
	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	redis "duolingo/libraries/work_distributor/drivers/redis"
	"duolingo/libraries/work_distributor/test/test_suites"
	"duolingo/test/fixtures"

	"github.com/stretchr/testify/suite"
)

func TestRedisWorkDistributor(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies("test", []string{
		"common",
		"connections",
	})

	provider := container.MustResolve[*facade.ConnectionProvider]()
	client := provider.GetRedisClient()
	distributor := redis.NewRedisWorkDistributor(client, 10)

	suite.Run(t, test_suites.NewWorkDistributorTestSuite(distributor))
}
