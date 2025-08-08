package dependencies

import (
	"context"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/services/user_service"
)

type UserService struct {
	connections *facade.ConnectionProvider
}

func NewUserService() *UserService {
	return &UserService{}
}

func (provider *UserService) Shutdown(shutdownCtx context.Context) {
}

func (provider *UserService) Bootstrap(bootstrapCtx context.Context, scope string) {
	provider.connections = container.MustResolve[*facade.ConnectionProvider]()

	container.BindSingleton[*user_service.UserService](func(ctx context.Context) any {
		return user_service.NewUserService(
			container.MustResolve[user_repo.UserRepoFactory](),
			container.MustResolve[user_repo.UserRepository](),
		)
	})
}
