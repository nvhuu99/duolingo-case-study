package providers

import (
	"context"
	container "duolingo/libraries/dependencies_container"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/services/user_service"
)

type UserServiceProvider struct {
}

func (provider *UserServiceProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *UserServiceProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	provider.registerMongoDBUserService()
}

func (provider *UserServiceProvider) registerMongoDBUserService() {
	container.BindSingleton[*user_service.UserService](func(ctx context.Context) any {
		return user_service.NewUserService(
			container.MustResolve[user_repo.UserRepoFactory](),
			container.MustResolve[user_repo.UserRepository](),
		)
	})
}
