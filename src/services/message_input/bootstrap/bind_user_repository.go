package bootstrap

import (
	"context"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/service_container"
	repo "duolingo/repositories/user_repository"
	repo_driver "duolingo/repositories/user_repository/drivers/mongodb"
)

func BindUserRepo() {
	container.BindSingleton[repo.UserRepository](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return repo_driver.NewUserRepo(provider.GetMongoClient(), "duolingo", "users")
	})
}
