package redis

import (
	"duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/work_distributor"
)

func NewRedisWorkDistributor(
	client *redis.RedisClient,
	distributionSize uint64,
) *work_distributor.WorkDistributor {
	proxy := NewRedisWorkStorageProxy(client)
	return work_distributor.NewWorkDistributor(proxy, 10)
}
