package mongodb

import (
	connection "duolingo/libraries/connection_manager/drivers/mongodb"
	mongo_cmd "duolingo/repositories/user_repository/drivers/mongodb/commands"
	mongo_services "duolingo/repositories/user_repository/drivers/mongodb/services"
	user_repo "duolingo/repositories/user_repository/external"
	cmd "duolingo/repositories/user_repository/external/commands"
	services "duolingo/repositories/user_repository/external/services"
	"sync"
)

type UserRepoFactory struct {
	client         *connection.MongoClient
	repo           user_repo.UserRepository
	createRepoOnce sync.Once
}

func NewUserRepoFactory(client *connection.MongoClient) *UserRepoFactory {
	return &UserRepoFactory{client: client}
}

func (factory *UserRepoFactory) MakeUserRepo() user_repo.UserRepository {
	factory.createRepoOnce.Do(func() {
		factory.repo = NewUserRepo(factory.client, "duolingo", "users")
	})
	return factory.repo
}

func (factory *UserRepoFactory) MakeUserService() services.UserService {
	return mongo_services.NewUserService(factory, factory.MakeUserRepo())
}

func (factory *UserRepoFactory) MakeListUsersCommand() cmd.ListUsersCommand {
	return mongo_cmd.NewListUsersCommand()
}

func (factory *UserRepoFactory) MakeDeleteUsersCommand() cmd.DeleteUsersCommand {
	return mongo_cmd.NewDeleteUsersCommand()
}

func (factory *UserRepoFactory) MakeAggregateUsersCommand() cmd.AggregateUsersCommand {
	return mongo_cmd.NewAggregateUsersCommand()
}
