package bootstrap

import (
	"context"

	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/service_container"
	"duolingo/libraries/work_distributor"
	redis "duolingo/libraries/work_distributor/drivers/redis"
)

func BindWorkDistributor() {
	container.BindSingleton[work_distributor.WorkDistributor](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return redis.NewRedisWorkDistributor(provider.GetRedisClient(), 10)
	})
}
