package rabbitmq

import (
	"context"

	"duolingo/libraries/connection_manager"
)

type RabbitMQConnectionBuilder struct {
	*connection_manager.ConnectionBuilder
	ctx context.Context
}

func NewRabbitMQConnectionBuilder(
	ctx context.Context,
	args *RabbitMQConnectionArgs,
) *RabbitMQConnectionBuilder {
	if args == nil {
		args = DefaultRabbitMQConnectionArgs()
	}
	baseBuilder := connection_manager.NewConnectionBuilder(ctx)
	baseBuilder.SetConnectionArgs(args)
	baseBuilder.SetConnectionProxy(&RabbitMQConnectionProxy{ctx: ctx})
	return &RabbitMQConnectionBuilder{
		ctx:               ctx,
		ConnectionBuilder: baseBuilder,
	}
}

func (builder *RabbitMQConnectionBuilder) BuildClientAndRegisterToManager() *RabbitMQClient {
	args, ok := builder.GetConnectionArgs().(*RabbitMQConnectionArgs)
	if !ok {
		panic(ErrConnectionArgsType)
	}
	client := builder.ConnectionBuilder.BuildClientAndRegisterToManager()
	rabbitMQClient := &RabbitMQClient{
		Client:                  client,
		declareTimeout: args.GetDeclareTimeout(),
	}

	return rabbitMQClient
}
