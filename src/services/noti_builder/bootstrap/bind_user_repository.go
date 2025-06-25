package bootstrap

import (
	"context"
	cnst "duolingo/constants"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/service_container"
	"duolingo/repositories/user_repository/drivers/mongodb"
)

func BindUserRepo() {
	container.BindSingletonAlias(cnst.UserRepoFactory, func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return mongodb.NewUserRepoFactory(provider.GetMongoClient())
	})

	container.BindSingletonAlias(cnst.UserRepo, func(ctx context.Context) any {
		factory := container.MustResolveAlias[*mongodb.UserRepoFactory](cnst.UserRepoFactory)
		return factory.MakeUserRepo()
	})

	container.BindSingletonAlias(cnst.UserService, func(ctx context.Context) any {
		factory := container.MustResolveAlias[*mongodb.UserRepoFactory](cnst.UserRepoFactory)
		return factory.MakeUserService()
	})
}
