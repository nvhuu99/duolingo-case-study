package distributed_lock

import (
	"context"
	"testing"
	"time"

	connection "duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/connection_manager/drivers/redis/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestMongoDBUserRepository(t *testing.T) {
	builder := connection.NewRedisConnectionBuilder(context.Background())
	builder.
		SetHost("localhost").
		SetPort("6379").
		SetLockAcquireTimeout(100*time.Millisecond).
		SetLockAcquireRetryWait(5*time.Millisecond, 10*time.Millisecond).
		SetLockTTL(1 * time.Second).
		SetConnectionTimeOut(100 * time.Millisecond).
		SetConnectionRetryWait(10 * time.Millisecond).
		SetOperationReadTimeOut(100 * time.Millisecond).
		SetOperationWriteTimeOut(100 * time.Millisecond).
		SetOperationRetryWait(10 * time.Millisecond)
	_, err := builder.BuildConnectionManager()
	if err != nil {
		panic(err)
	}
	client, err := builder.BuildClientAndRegisterToManager()
	if err != nil {
		panic(err)
	}

	suite.Run(t, test_suites.NewDistributedLockTestSuite(client))
}
