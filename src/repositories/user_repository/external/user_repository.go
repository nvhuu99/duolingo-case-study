package external

import (
	"duolingo/models"
	"duolingo/repositories/user_repository/external/commands"
	"duolingo/repositories/user_repository/external/commands/results"
)

type UserRepository interface {
	InsertManyUsers(users []*models.User) ([]*models.User, error)

	DeleteUsersByIds(ids []string) error
	DeleteUsers(cmd commands.DeleteUsersCommand) error

	GetListUsersByIds(ids []string) ([]*models.User, error)
	GetListUsers(cmd commands.ListUsersCommand) ([]*models.User, error)
	GetListUserDevices(cmd commands.ListUserDevicesCommand) ([]*models.UserDevice, error)

	AggregateUsers(cmd commands.AggregateUsersCommand) (results.UsersAggregationResult, error)
}
