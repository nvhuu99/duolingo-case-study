package command_builders

import (
	"time"

	b "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserDevicesAggregationCommandBuilder struct {
	filters       b.M
	pipelineSteps []b.D
	pipeline      mongo.Pipeline
}

func (opts *UserDevicesAggregationCommandBuilder) WithCampaignFilter(campaign string) *UserDevicesAggregationCommandBuilder {
	if opts.filters == nil {
		opts.filters = b.M{}
	}
	opts.filters["campaigns"] = campaign
	return opts
}

func (opts *UserDevicesAggregationCommandBuilder) WithEmailVerifiedOnlyFilter() *UserDevicesAggregationCommandBuilder {
	if opts.filters == nil {
		opts.filters = b.M{}
	}
	opts.filters["verified_at"] = b.M{"$lte": time.Now()}
	return opts
}

func (opts *UserDevicesAggregationCommandBuilder) Build() error {
	var pipeline mongo.Pipeline
	if len(opts.filters) > 0 {
		pipeline = append(pipeline, b.D{{Key: "$match", Value: opts.filters}})
	}
	for i := range opts.pipelineSteps {
		pipeline = append(pipeline, opts.pipelineSteps[i])
	}
	opts.pipeline = pipeline
	return nil
}

func (opts *UserDevicesAggregationCommandBuilder) WithSumUserDevicesAggregation() *UserDevicesAggregationCommandBuilder {
	opts.pipelineSteps = append(opts.pipelineSteps,
		b.D{{Key: "$project", Value: b.M{
			"count": b.M{"$size": "$device_tokens"},
		}}},
		b.D{{Key: "$group", Value: b.M{
			"_id":   nil,
			"total": b.M{"$sum": "$count"},
		}}},
	)
	return opts
}

func (opts *UserDevicesAggregationCommandBuilder) GetPipeline() mongo.Pipeline {
	return opts.pipeline
}
