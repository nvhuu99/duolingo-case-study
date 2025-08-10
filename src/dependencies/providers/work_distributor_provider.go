package providers

import (
	"context"

	"duolingo/libraries/config_reader"
	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	dist "duolingo/libraries/work_distributor"
	redis "duolingo/libraries/work_distributor/drivers/redis"
)

type WorkDistributorProvider struct {
}

func (provider *WorkDistributorProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	provider.registerRedisWorkDistributor()
}

func (provider *WorkDistributorProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *WorkDistributorProvider) registerRedisWorkDistributor() {
	container.BindSingleton[*dist.WorkDistributor](func(ctx context.Context) any {
		config := container.MustResolve[config_reader.ConfigReader]()
		connections := container.MustResolve[*facade.ConnectionProvider]()
		return redis.NewRedisWorkDistributor(
			connections.GetRedisClient(),
			config.GetInt64("work_distributor", "distribution_size"),
		)
	})
}
