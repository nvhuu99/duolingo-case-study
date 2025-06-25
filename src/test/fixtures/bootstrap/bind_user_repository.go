package bootstrap

import (
	"context"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/service_container"
	"duolingo/repositories/user_repository/drivers/mongodb"
	cnst "duolingo/test/fixtures/constants"
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
