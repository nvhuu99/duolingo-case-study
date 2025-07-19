package dependencies

import (
	"context"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	"duolingo/repositories/user_repository/drivers/mongodb"
	user_repo "duolingo/repositories/user_repository/external"
)

type UserRepo struct {
	connections *facade.ConnectionProvider
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		connections: container.MustResolve[*facade.ConnectionProvider](),
	}
}

func (provider *UserRepo) Shutdown() {
}

func (provider *UserRepo) Bootstrap() {
	container.BindSingleton[user_repo.UserRepoFactory](func(ctx context.Context) any {
		return mongodb.NewUserRepoFactory(provider.connections.GetMongoClient())
	})

	container.BindSingleton[user_repo.UserRepository](func(ctx context.Context) any {
		factory := container.MustResolve[user_repo.UserRepoFactory]()
		return factory.MakeUserRepo()
	})
}
