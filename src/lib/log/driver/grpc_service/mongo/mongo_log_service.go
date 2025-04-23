package mongo

import (
	"context"
	lw "duolingo/lib/log/writer"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoLogService struct {
	Database   string
	Collection string

	ctx context.Context

	mongoClient *mongo.Client
}

func NewMongoLogService(ctx context.Context, uri string, db string, coll string) (*MongoLogService, error) {
	sv := &MongoLogService{
		ctx:        ctx,
		Database:   db,
		Collection: coll,
	}

	opts := options.Client()
	opts.SetConnectTimeout(30 * time.Second)
	opts.SetSocketTimeout(10 * time.Second)
	opts.ApplyURI(uri)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	sv.mongoClient = client

	return sv, nil
}

func (srv *MongoLogService) Write(line *lw.Writable) error {
	var data any
	if err := json.Unmarshal(line.Content, &data); err != nil {
		return err
	}
	collection := srv.mongoClient.Database(srv.Database).Collection(srv.Collection)
	if _, err := collection.InsertOne(srv.ctx, data); err != nil {
		return err
	}
	return nil
}
