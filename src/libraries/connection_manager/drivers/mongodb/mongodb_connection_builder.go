package mongodb

import (
	"context"

	"duolingo/libraries/connection_manager"
)

type MongoConnectionBuilder struct {
	*connection_manager.ConnectionBuilder
	ctx context.Context
}

func NewMongoConnectionBuilder(
	ctx context.Context,
	args *MongoConnectionArgs,
) *MongoConnectionBuilder {
	if args == nil {
		args = DefaultMongoConnectionArgs()
	}
	baseBuilder := connection_manager.NewConnectionBuilder(ctx)
	baseBuilder.SetConnectionArgs(args)
	baseBuilder.SetConnectionProxy(&MongoConnectionProxy{ctx: ctx})
	return &MongoConnectionBuilder{
		ctx:               ctx,
		ConnectionBuilder: baseBuilder,
	}
}

func (builder *MongoConnectionBuilder) BuildClientAndRegisterToManager() *MongoClient {
	client := builder.ConnectionBuilder.BuildClientAndRegisterToManager()
	mongoClient := &MongoClient{
		Client: client,
	}

	return mongoClient
}
