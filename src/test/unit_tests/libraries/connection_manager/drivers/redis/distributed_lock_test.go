package distributed_lock

import (
	"context"
	"testing"
	"time"

	"duolingo/dependencies"
	"duolingo/libraries/config_reader"
	redis "duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/connection_manager/drivers/redis/test/test_suites"
	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	"duolingo/test/fixtures"

	"github.com/stretchr/testify/suite"
)

func TestDistributedLock(t *testing.T) {
	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies("test", []string{
		"common",
	})

	config := container.MustResolve[config_reader.ConfigReader]()
	provider := facade.Provider(context.Background())
	provider.InitRedis(redis.
		DefaultRedisConnectionArgs().
		SetLockAcquireTimeout(500 * time.Millisecond).
		SetLockTTL(2 * time.Second).
		SetHost(config.Get("redis", "host")).
		SetPort(config.Get("redis", "port")),
	)

	suite.Run(t, test_suites.NewDistributedLockTestSuite(provider.GetRedisClient()))
}
