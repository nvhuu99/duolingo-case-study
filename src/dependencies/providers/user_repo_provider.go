package providers

import (
	"context"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	"duolingo/repositories/user_repository/drivers/mongodb"
	user_repo "duolingo/repositories/user_repository/external"
)

type UserRepoProvider struct {
}

func (provider *UserRepoProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *UserRepoProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	provider.registerMongoDBUserRepo()
}

func (provider *UserRepoProvider) registerMongoDBUserRepo() {
	container.BindSingleton[user_repo.UserRepoFactory](func(ctx context.Context) any {
		connections := container.MustResolve[*facade.ConnectionProvider]()
		return mongodb.NewUserRepoFactory(connections.GetMongoClient())
	})

	container.BindSingleton[user_repo.UserRepository](func(ctx context.Context) any {
		factory := container.MustResolve[user_repo.UserRepoFactory]()
		return factory.MakeUserRepo()
	})
}
