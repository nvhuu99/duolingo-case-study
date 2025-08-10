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

func (proxy *MongoConnectionProxy) ConnectionName() string { return "MongoDB" }

func (proxy *MongoConnectionProxy) SetArgsPanicIfInvalid(args any) {
	mongoArgs, ok := args.(*MongoConnectionArgs)
	if !ok {
		panic(ErrConnectionArgsType)
	}
	if mongoArgs.GetHost() == "" || mongoArgs.GetPort() == "" {
		panic(ErrInvalidConnectionArgs)
	}
	if mongoArgs.GetURI() == "" {
		address := fmt.Sprintf("%v:%v", mongoArgs.GetHost(), mongoArgs.GetPort())
		credentials := fmt.Sprintf(
			"%v:%v", 
			mongoArgs.GetUser(), 
			mongoArgs.GetPassword(),
		)
		uri := fmt.Sprintf("mongodb://%v/", address)
		if credentials != ":" {
			uri = fmt.Sprintf("mongodb://%v@%v/", credentials, address)
		}
		fmt.Println(uri)
		mongoArgs.SetURI(uri)
	}
	proxy.connectionArgs = mongoArgs
}

func (proxy *MongoConnectionProxy) MakeConnection() (any, error) {
	args := proxy.connectionArgs
	opts := options.Client()
	opts.SetConnectTimeout(args.GetConnectionTimeout())
	opts.ApplyURI(args.GetURI())
	return mongo.Connect(proxy.ctx, opts)
}

func (proxy *MongoConnectionProxy) Ping(connection any) error {
	if mongoConn, ok := connection.(*mongo.Client); ok {
		return mongoConn.Ping(proxy.ctx, nil)
	}
	return ErrConnectionType
}

func (proxy *MongoConnectionProxy) IsNetworkErr(err error) bool {
	return mongo.IsNetworkError(err)
}

func (proxy *MongoConnectionProxy) CloseConnection(connection any) {
	if mongoConn, ok := connection.(*mongo.Client); ok {
		mongoConn.Disconnect(proxy.ctx)
	}
}
