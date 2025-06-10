package redis

import (
	"context"
	"testing"
	"time"

	connection "duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/work_distributor"
	redis_driver "duolingo/libraries/work_distributor/drivers/redis"
	"duolingo/libraries/work_distributor/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestRedisWorkDistributor(t *testing.T) {
	builder := connection.NewRedisConnectionBuilder(context.Background())
	builder.
		SetHost("localhost").
		SetPort("6379").
		SetLockAcquireTimeout(200*time.Millisecond).
		SetLockAcquireRetryWait(2*time.Millisecond, 5*time.Millisecond).
		SetLockTTL(100 * time.Second).
		SetConnectionTimeOut(50 * time.Second).
		SetConnectionRetryWait(5 * time.Millisecond).
		SetOperationReadTimeOut(50 * time.Millisecond).
		SetOperationWriteTimeOut(50 * time.Millisecond).
		SetOperationRetryWait(5 * time.Millisecond)
	defer builder.Destroy()

	_, err := builder.BuildConnectionManager()
	if err != nil {
		panic(err)
	}

	client, err := builder.BuildClientAndRegisterToManager()
	if err != nil {
		panic(err)
	}

	proxy := redis_driver.NewRedisWorkStorageProxy(client)

	distributor := work_distributor.NewWorkDistributor(proxy, 10)

	suite.Run(t, test_suites.NewWorkDistributorTestSuite(distributor))
}
