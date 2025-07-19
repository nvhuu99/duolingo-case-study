package external

import (
	"duolingo/repositories/user_repository/external/commands"
)

type UserRepoFactory interface {
	MakeUserRepo() UserRepository
	MakeListUsersCommand() commands.ListUsersCommand
	MakeListUserDevicesCommand() commands.ListUserDevicesCommand
	MakeDeleteUsersCommand() commands.DeleteUsersCommand
	MakeAggregateUsersCommand() commands.AggregateUsersCommand
}
