package dependencies

import (
	"context"

	"duolingo/libraries/config_reader"
	facade "duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	dist "duolingo/libraries/work_distributor"
	redis "duolingo/libraries/work_distributor/drivers/redis"
)

type WorkDistributor struct {
	config      config_reader.ConfigReader
	connections *facade.ConnectionProvider
}

func NewWorkDistributor() *WorkDistributor {
	return &WorkDistributor{}
}

func (provider *WorkDistributor) Bootstrap(bootstrapCtx context.Context, scope string) {
	provider.config = container.MustResolve[config_reader.ConfigReader]()
	provider.connections = container.MustResolve[*facade.ConnectionProvider]()

	container.BindSingleton[*dist.WorkDistributor](func(ctx context.Context) any {
		return redis.NewRedisWorkDistributor(
			provider.connections.GetRedisClient(),
			provider.config.GetInt64("work_distributor", "distribution_size"),
		)
	})
}

func (provider *WorkDistributor) Shutdown(shutdownCtx context.Context) {
}
