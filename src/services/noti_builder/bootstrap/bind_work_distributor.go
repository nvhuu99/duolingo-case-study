package bootstrap

import (
	"context"

	facade "duolingo/libraries/connection_manager/facade"
	ps "duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/libraries/work_distributor"
	redis "duolingo/libraries/work_distributor/drivers/redis"
	"duolingo/repositories/user_repository/external/services"
	wrkl "duolingo/services/noti_builder/server/workloads"
	cnst "duolingo/constants"
)

func BindWorkDistributor() {
	container.BindSingleton[*work_distributor.WorkDistributor](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return redis.NewRedisWorkDistributor(provider.GetRedisClient(), cnst.DistributionSize)
	})

	container.BindSingleton[*wrkl.TokenBatchDistributor](func(ctx context.Context) any {
		userService := container.MustResolveAlias[services.UserService](cnst.UserService)
		distributor := container.MustResolve[*work_distributor.WorkDistributor]()
		jobPublisher := container.MustResolveAlias[ps.Publisher](cnst.NotiBuilderJobPublisher)
		jobSubscriber := container.MustResolveAlias[ps.Subscriber](cnst.NotiBuilderJobSubscriber)
		return wrkl.NewTokenBatchDistributor(
			distributor,
			jobPublisher,
			jobSubscriber,
			userService,
		)
	})
}
