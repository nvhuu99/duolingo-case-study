package mongodb

import (
	"context"
	"net/url"

	"duolingo/libraries/connection_manager"
)

type MongoConnectionBuilder struct {
	connection_manager.ConnectionBuilder

	driver *MongoConnectionProxy

	Host    string
	Port    string
	User    string
	Passwod string

	ctx context.Context
}

func NewMongoConnectionBuilder(ctx context.Context) *MongoConnectionBuilder {
	builder := &MongoConnectionBuilder{
		ctx: ctx,
	}
	builder.ConnectionBuilder = *connection_manager.NewConnectionBuilder(ctx)
	builder.driver = &MongoConnectionProxy{ctx: ctx}
	return builder
}

func (builder *MongoConnectionBuilder) SetHost(host string) *MongoConnectionBuilder {
	builder.Host = host
	return builder
}

func (builder *MongoConnectionBuilder) SetPort(port string) *MongoConnectionBuilder {
	builder.Port = port
	return builder
}

func (builder *MongoConnectionBuilder) SetCredentials(user string, password string) *MongoConnectionBuilder {
	builder.User = url.QueryEscape(user)
	builder.Passwod = url.QueryEscape(password)
	return builder
}

func (builder *MongoConnectionBuilder) BuildConnectionManager() (*connection_manager.ConnectionManager, error) {
	args := &MongoConnectionArgs{
		ConnectionArgs: connection_manager.ConnectionArgs{
			ConnectionTimeout:     builder.ConnectionTimeout,
			ConnectionRetryWait:   builder.ConnectionRetryWait,
			OperationRetryWait:    builder.OperationRetryWait,
			OperationReadTimeout:  builder.OperationReadTimeout,
			OperationWriteTimeout: builder.OperationWriteTimeout,
		},
		Host:     builder.Host,
		Port:     builder.Port,
		User:     builder.User,
		Password: builder.Passwod,
	}
	builder.driver.SetConnectionArgsWithPanicOnValidationErr(args)
	builder.SetConnectionDriver(builder.driver)

	return builder.ConnectionBuilder.BuildConnectionManager()
}

func (builder *MongoConnectionBuilder) BuildClientAndRegisterToManager() (*MongoClient, error) {
	client, err := builder.ConnectionBuilder.BuildClientAndRegisterToManager()
	if err != nil {
		return nil, err
	}

	mongoClient := &MongoClient{
		Client: *client,
	}

	return mongoClient, nil
}
