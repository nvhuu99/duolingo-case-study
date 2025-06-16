package mongodb

import (
	connection "duolingo/libraries/connection_manager/drivers/mongodb"
	mongo_cmd "duolingo/repositories/user_repository/drivers/mongodb/commands"
	user_repo "duolingo/repositories/user_repository/external"
	cmd "duolingo/repositories/user_repository/external/commands"
)

type UserRepoFactory struct {
	client *connection.MongoClient
}

func NewUserRepoFactory(client *connection.MongoClient) *UserRepoFactory {
	return &UserRepoFactory{client: client}
}

func (factory *UserRepoFactory) MakeUserRepo() user_repo.UserRepository {
	return NewUserRepo(factory.client, "duolingo", "users")
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
