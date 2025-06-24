package commands

import (
	"duolingo/repositories/user_repository/drivers/mongodb/commands/filters"

	cmd "duolingo/repositories/user_repository/external/commands"
	b "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ListUserDevicesCommand struct {
	*filters.UserFilters

	pipeline mongo.Pipeline
	sorts    b.M
	offset   uint64
	limit    uint64
}

func NewListUserDevicesCommand() *ListUserDevicesCommand {
	return &ListUserDevicesCommand{
		UserFilters: filters.NewUserFilters(),
		pipeline:    mongo.Pipeline{},
		sorts:       b.M{},
	}
}

func (command *ListUserDevicesCommand) SetPagination(offset uint64, limit uint64) {
	command.offset = offset
	command.limit = limit
}

func (command *ListUserDevicesCommand) SetSortById(ord cmd.SortOrder) {
	if ord == cmd.OrderASC {
		command.sorts["user_id"] = 1
	} else {
		command.sorts["user_id"] = -1
	}
}

func (command *ListUserDevicesCommand) Build() error {
	command.pipeline = mongo.Pipeline{
		{{Key: "$match", Value: command.GetFilters()}},
		{{Key: "$unwind", Value: b.M{"path": "$user_devices"}}},
		{{Key: "$sort", Value: command.sorts}},
		{{Key: "$project", Value: b.M{
			"platform": "$user_devices.platform",
			"token":    "$user_devices.token",
		}}},
		{{Key: "$skip", Value: command.offset}},
		{{Key: "$limit", Value: command.limit}},
	}
	return nil
}

func (command *ListUserDevicesCommand) GetPipeline() mongo.Pipeline {
	return command.pipeline
}
