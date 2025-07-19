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
	return &UserService{
		connections: container.MustResolve[*facade.ConnectionProvider](),
	}
}

func (provider *UserService) Shutdown() {
}

func (provider *UserService) Bootstrap() {
	container.BindSingleton[user_service.UserService](func(ctx context.Context) any {
		return user_service.NewUserService(
			container.MustResolve[user_repo.UserRepoFactory](),
			container.MustResolve[user_repo.UserRepository](),
		)
	})
}
