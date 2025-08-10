package external

import (
	"context"
	"duolingo/models"
	"duolingo/repositories/user_repository/external/commands"
	"duolingo/repositories/user_repository/external/commands/results"
)

type UserRepository interface {
	InsertManyUsers(ctx context.Context, users []*models.User) ([]*models.User, error)

	DeleteUsersByIds(ctx context.Context, ids []string) error
	DeleteUsers(ctx context.Context, cmd commands.DeleteUsersCommand) error

	GetListUsersByIds(ctx context.Context, ids []string) ([]*models.User, error)
	GetListUsers(ctx context.Context, cmd commands.ListUsersCommand) ([]*models.User, error)
	GetListUserDevices(ctx context.Context, cmd commands.ListUserDevicesCommand) ([]*models.UserDevice, error)

	AggregateUsers(ctx context.Context, cmd commands.AggregateUsersCommand) (results.UsersAggregationResult, error)
}
