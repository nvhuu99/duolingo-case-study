package command_builders

import (
	b "go.mongodb.org/mongo-driver/bson"
)

type UserDeleteCommandBuilder struct {
	filters b.M
}

func (opts *UserDeleteCommandBuilder) WithUserIdsFilter(ids []string) *UserDeleteCommandBuilder {
	if opts.filters == nil {
		opts.filters = b.M{}
	}
	opts.filters["user_id"] = b.M{"$in": ids}
	return opts
}

func (opts *UserDeleteCommandBuilder) WithCampaignFilter(campaign string) *UserDeleteCommandBuilder {
	if opts.filters == nil {
		opts.filters = b.M{}
	}
	opts.filters["campaigns"] = campaign
	return opts
}

func (opts *UserDeleteCommandBuilder) Build() error {
	return nil
}

func (opts *UserDeleteCommandBuilder) GetFilters() b.M {
	return opts.filters
}
