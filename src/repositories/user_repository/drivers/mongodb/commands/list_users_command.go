package commands

import (
	"duolingo/repositories/user_repository/drivers/mongodb/commands/filters"
	cmd "duolingo/repositories/user_repository/external/commands"

	b "go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo/options"
)

type ListUsersCommand struct {
	*filters.UserFilters
	options *mongo.FindOptions
	sorts   b.M
}

func NewListUsersCommand() *ListUsersCommand {
	return &ListUsersCommand{
		UserFilters: filters.NewUserFilters(),
		options:     &mongo.FindOptions{},
		sorts:       b.M{},
	}
}

func (command *ListUsersCommand) SetPagination(offset int64, limit int64) {
	command.options.SetSkip(int64(offset))
	command.options.SetLimit(int64(limit))
}

func (command *ListUsersCommand) SetSortById(ord cmd.SortOrder) {
	if ord == cmd.OrderASC {
		command.sorts["user_id"] = 1
	} else {
		command.sorts["user_id"] = -1
	}
}

func (command *ListUsersCommand) Build() error {
	if len(command.sorts) > 0 {
		command.options.SetSort(command.sorts)
	}
	return nil
}

func (command *ListUsersCommand) GetOptions() *mongo.FindOptions {
	return command.options
}
