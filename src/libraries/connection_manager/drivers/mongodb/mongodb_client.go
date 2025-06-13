package mongodb

import (
	"context"
	"duolingo/libraries/connection_manager"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoClient struct {
	*connection_manager.Client
}

func (client *MongoClient) ExecuteClosure(
	wait time.Duration,
	closure func(ctx context.Context, connection *mongo.Client) error,
) error {
	wrapper := func(ctx context.Context, conn any) error {
		converted, _ := conn.(*mongo.Client)
		return closure(ctx, converted)
	}
	return client.Client.ExecuteClosure(wait, wrapper)
}

func (client *MongoClient) GetConnection() *mongo.Client {
	connection := client.Client.GetConnection()
	if mongoConn, ok := connection.(*mongo.Client); ok {
		return mongoConn
	}
	return nil
}
