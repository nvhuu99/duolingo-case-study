package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrConnectionType        = errors.New("provided with a connection not mongo.Client")
	ErrConnectionArgsType    = errors.New("provided with an argument type that is not MongoConnectionArgs")
	ErrInvalidConnectionArgs = errors.New("provided with invalid connection arguments")
)

type MongoConnectionProxy struct {
	ctx            context.Context
	connectionArgs *MongoConnectionArgs
}

/* Implement connection_manager.ConnectionProxy interface */

func (proxy *MongoConnectionProxy) SetConnectionArgsWithPanicOnValidationErr(args any) {
	mongoArgs, ok := args.(*MongoConnectionArgs)
	if !ok {
		panic(ErrConnectionArgsType)
	}
	if mongoArgs.Host == "" || mongoArgs.Port == "" {
		panic(ErrInvalidConnectionArgs)
	}
	if mongoArgs.URI == "" {
		address := fmt.Sprintf("%v:%v", mongoArgs.Host, mongoArgs.Port)
		credentials := fmt.Sprintf("%v:%v", mongoArgs.User, mongoArgs.Password)
		uri := fmt.Sprintf("mongodb://%v/", address)
		if credentials != ":" {
			uri = fmt.Sprintf("mongodb://%v@%v/", credentials, address)
		}
		mongoArgs.URI = uri
	}
	proxy.connectionArgs = mongoArgs
}

func (proxy *MongoConnectionProxy) CreateConnection() (any, error) {
	args := proxy.connectionArgs
	opts := options.Client()
	opts.SetConnectTimeout(args.ConnectionTimeout)
	opts.ApplyURI(args.URI)
	return mongo.Connect(proxy.ctx, opts)
}

func (proxy *MongoConnectionProxy) Ping(connection any) error {
	if mongoConn, ok := connection.(*mongo.Client); ok {
		return mongoConn.Ping(proxy.ctx, nil)
	}
	return ErrConnectionType
}

func (proxy *MongoConnectionProxy) IsNetworkError(err error) bool {
	return mongo.IsNetworkError(err)
}

func (proxy *MongoConnectionProxy) CloseConnection(connection any) {
	if mongoConn, ok := connection.(*mongo.Client); ok {
		mongoConn.Disconnect(proxy.ctx)
	}
}
