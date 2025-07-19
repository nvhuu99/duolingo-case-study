package mongodb

import (
	"sync"

	connection "duolingo/libraries/connection_manager/drivers/mongodb"
	mongo_cmd "duolingo/repositories/user_repository/drivers/mongodb/commands"
	user_repo "duolingo/repositories/user_repository/external"
	cmd "duolingo/repositories/user_repository/external/commands"
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

func (factory *UserRepoFactory) MakeListUsersCommand() cmd.ListUsersCommand {
	return mongo_cmd.NewListUsersCommand()
}

func (factory *UserRepoFactory) MakeListUserDevicesCommand() cmd.ListUserDevicesCommand {
	return mongo_cmd.NewListUserDevicesCommand()
}

func (factory *UserRepoFactory) MakeDeleteUsersCommand() cmd.DeleteUsersCommand {
	return mongo_cmd.NewDeleteUsersCommand()
}

func (factory *UserRepoFactory) MakeAggregateUsersCommand() cmd.AggregateUsersCommand {
	return mongo_cmd.NewAggregateUsersCommand()
}
