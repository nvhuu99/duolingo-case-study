package commands

import (
	"duolingo/repositories/user_repository/drivers/mongodb/commands/filters"

	b "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AggregateUsersCommand struct {
	*filters.UserFilters

	pipelineSteps []b.D
	pipeline      mongo.Pipeline
}

func NewAggregateUsersCommand() *AggregateUsersCommand {
	return &AggregateUsersCommand{
		UserFilters:   filters.NewUserFilters(),
		pipelineSteps: []b.D{},
		pipeline:      mongo.Pipeline{},
	}
}

func (command *AggregateUsersCommand) AddAggregationSumUserDevices() {
	command.pipelineSteps = append(command.pipelineSteps,
		b.D{{Key: "$project", Value: b.M{
			"count": b.M{"$size": "$device_tokens"},
		}}},
		b.D{{Key: "$group", Value: b.M{
			"_id":                nil,
			"count_user_devices": b.M{"$sum": "$count"},
		}}},
	)
}

func (command *AggregateUsersCommand) Build() error {
	command.pipeline = mongo.Pipeline{}

	filters := command.GetFilters()
	if len(filters) > 0 {
		command.pipeline = append(command.pipeline, b.D{{Key: "$match", Value: filters}})
	}

	for i := range command.pipelineSteps {
		command.pipeline = append(command.pipeline, command.pipelineSteps[i])
	}

	return nil
}

func (command *AggregateUsersCommand) GetPipeline() mongo.Pipeline {
	return command.pipeline
}
