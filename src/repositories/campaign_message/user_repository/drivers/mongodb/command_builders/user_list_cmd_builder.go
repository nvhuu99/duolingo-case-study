package command_builders

import (
	b "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserListCommandBuilder struct {
	base    *options.FindOptions
	filters b.M
	sorts   b.D
}

func (opts *UserListCommandBuilder) AddSort(key string, ord SortOrder) *UserListCommandBuilder {
	opts.sorts = append(opts.sorts, b.E{Key: key, Value: ord})
	return opts
}

func (opts *UserListCommandBuilder) WithUserIdsFilter(ids []string) *UserListCommandBuilder {
	if opts.filters == nil {
		opts.filters = b.M{}
	}
	opts.filters["user_id"] = b.M{"$in": ids}
	return opts
}

func (opts *UserListCommandBuilder) WithCampaignFilter(campaign string) *UserListCommandBuilder {
	if opts.filters == nil {
		opts.filters = b.M{}
	}
	opts.filters["campaigns"] = campaign
	return opts
}

func (opts *UserListCommandBuilder) Build() error {
	if len(opts.sorts) > 0 {
		opts.base.SetSort(opts.sorts)
	}
	return nil
}

func (opts *UserListCommandBuilder) GetFilters() b.M {
	return opts.filters
}

func (opts *UserListCommandBuilder) GetOptions() *options.FindOptions {
	return opts.base
}
