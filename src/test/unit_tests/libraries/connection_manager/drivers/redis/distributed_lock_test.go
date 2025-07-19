package distributed_lock

import (
	"context"
	"testing"
	"time"

	redis "duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/connection_manager/drivers/redis/test/test_suites"
	facade "duolingo/libraries/connection_manager/facade"

	"github.com/stretchr/testify/suite"
)

func TestDistributedLock(t *testing.T) {
	client := facade.Provider(context.Background()).InitRedis(redis.
		DefaultRedisConnectionArgs().
		SetLockAcquireTimeout(500 * time.Millisecond).
		SetLockTTL(2 * time.Second),
	).GetRedisClient()

	suite.Run(t, test_suites.NewDistributedLockTestSuite(client))
}
