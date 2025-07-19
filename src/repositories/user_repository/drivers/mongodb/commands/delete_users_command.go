package commands

import (
	"duolingo/repositories/user_repository/drivers/mongodb/commands/filters"
)

type DeleteUsersCommand struct {
	*filters.UserFilters
}

func NewDeleteUsersCommand() *DeleteUsersCommand {
	return &DeleteUsersCommand{
		UserFilters: filters.NewUserFilters(),
	}
}

func (command *DeleteUsersCommand) Build() error {
	return nil
}
