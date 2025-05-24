package mongo

import (
	"context"
	lw "duolingo/lib/log/writer"
	"encoding/json"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoLogService struct {
	Database   string
	Collection string

	ctx             context.Context
	bufferCh        chan any
	bufferSizeLimit int

	mu        sync.Mutex
	lastErr   error

	mongoClient *mongo.Client
}

func NewMongoLogService(ctx context.Context, uri string, db string, coll string) (*MongoLogService, error) {
	sv := &MongoLogService{
		ctx:             ctx,
		Database:        db,
		Collection:      coll,
		bufferSizeLimit: 200,
		bufferCh:        make(chan any, 200),
	}

	opts := options.Client().
		SetConnectTimeout(30 * time.Second).
		SetSocketTimeout(10 * time.Second).
		ApplyURI(uri)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	sv.mongoClient = client
	go sv.buffering()

	return sv, nil
}

func (srv *MongoLogService) Write(line *lw.Writable) error {
	var data any
	if err := json.Unmarshal(line.Content, &data); err != nil {
		return err
	}
	srv.bufferCh <- data
	return srv.consumeLastErr()
}

func (srv *MongoLogService) buffering() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	batch := make([]any, 0, srv.bufferSizeLimit)

	write := func(batch []any) error {
		if len(batch) == 0 {
			return nil
		}
		collection := srv.mongoClient.Database(srv.Database).Collection(srv.Collection)
		_, err := collection.InsertMany(srv.ctx, batch)
		if err == nil {
			log.Printf("mongo log service: insert batch len(%v)\n", len(batch))
		}
		return err
	}

	for {
		select {
		case <-srv.ctx.Done():
			err := write(batch)
			srv.setLastErr(err)
			return

		case <-ticker.C:
			err := write(batch)
			srv.setLastErr(err)
			batch = batch[:0] // clear a slice without reallocating

		case line := <-srv.bufferCh:
			batch = append(batch, line)
			if len(batch) >= srv.bufferSizeLimit {
				err := write(batch)
				srv.setLastErr(err)
				batch = batch[:0]
			}
		}
	}
}

func (srv *MongoLogService) consumeLastErr() error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	err := srv.lastErr
	srv.lastErr = nil
	return err
}

func (srv *MongoLogService) setLastErr(err error) {
	if err == nil {
		return
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.lastErr == nil {
		srv.lastErr = err
	}
}
