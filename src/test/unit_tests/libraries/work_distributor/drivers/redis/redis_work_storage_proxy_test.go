package redis

import (
	"context"
	"testing"

	facade "duolingo/libraries/connection_manager/facade"
	redis "duolingo/libraries/work_distributor/drivers/redis"
	"duolingo/libraries/work_distributor/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestRedisWorkStorageProxy(t *testing.T) {
	client := facade.Provider(context.Background()).InitRedis(nil).GetRedisClient()
	proxy := redis.NewRedisWorkStorageProxy(client)

	suite.Run(t, test_suites.NewWorkStorageProxyTestSuite(proxy))
}
