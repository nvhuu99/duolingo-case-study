package bootstrap

import (
	"context"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/service_container"
	repo_driver "duolingo/repositories/user_repository/drivers/mongodb"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/repositories/user_repository/external/services"
)

func BindUserRepo() {
	container.BindSingleton[user_repo.UserRepoFactory](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return repo_driver.NewUserRepoFactory(provider.GetMongoClient())
	})

	container.BindSingleton[user_repo.UserRepository](func(ctx context.Context) any {
		factory := container.MustResolve[user_repo.UserRepoFactory]()
		return factory.MakeUserRepo()
	})

	container.BindSingleton[services.UserService](func(ctx context.Context) any {
		factory := container.MustResolve[user_repo.UserRepoFactory]()
		return factory.MakeUserService()
	})
}
