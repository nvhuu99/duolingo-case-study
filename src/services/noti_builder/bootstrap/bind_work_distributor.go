package bootstrap

import (
	"context"

	facade "duolingo/libraries/connection_manager/facade"
	"duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/libraries/work_distributor"
	redis "duolingo/libraries/work_distributor/drivers/redis"
	"duolingo/repositories/user_repository/external/services"
	"duolingo/services/noti_builder/server/workloads"
)

func BindWorkDistributor() {
	container.BindSingleton[*work_distributor.WorkDistributor](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return redis.NewRedisWorkDistributor(provider.GetRedisClient(), 10)
	})

	container.BindSingleton[*workloads.TokenBatchDistributor](func(ctx context.Context) any {
		userService := container.MustResolve[services.UserService]()
		distributor := container.MustResolve[*work_distributor.WorkDistributor]()
		publisher := container.MustResolve[pub_sub.Publisher]()
		return workloads.NewTokenBatchDistributor(
			distributor,
			userService,
			publisher,
		)
	})
}
