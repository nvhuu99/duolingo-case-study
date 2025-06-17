package external

import (
	"duolingo/repositories/user_repository/external/commands"
	"duolingo/repositories/user_repository/external/services"
)

type UserRepoFactory interface {
	MakeUserRepo() UserRepository
	MakeUserService() services.UserService
	MakeListUsersCommand() commands.ListUsersCommand
	MakeDeleteUsersCommand() commands.DeleteUsersCommand
	MakeAggregateUsersCommand() commands.AggregateUsersCommand
}
