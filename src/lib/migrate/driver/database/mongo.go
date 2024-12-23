package mongo_driver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	migrate "duolingo/lib/migrate"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectionTimeOut = 10 * time.Second
	defaultTimeOut    = 30 * time.Second
)

var (
	ErrPrepareDatabse     = "Mongo: failed to create \"migrations\" table"
	ErrInvalidJsonCommand = "Mongo: failed to parse json command"
)

type Mongo struct {
	conStr   string
	database string
	timeOut  time.Duration
	ctx      context.Context
}

func New(ctx context.Context, conStr string, database string) *Mongo {
	driver := Mongo{}
	driver.ctx = ctx
	driver.timeOut = defaultTimeOut
	driver.conStr = conStr
	driver.database = database

	return &driver
}

func (driver *Mongo) SetOperationTimeOut(duration time.Duration) *Mongo {
	driver.timeOut = duration
	return driver
}

func (driver *Mongo) PrepareDatabase() error {
	listResult, err := driver.executeCommand(`{
		"listCollections": 1,
		"filter": { "name", "migrations" },
		"nameOnly": true
	}`)

	if err != nil {
		return err
	}

	if len(listResult) == 0 {
		if _, err := driver.executeCommand(`{ "create": "migrations" }`); err != nil {
			return errors.New(ErrPrepareDatabse)
		}
		_, err = driver.executeCommand(`{
			"createIndexes": "migrations",
			"indexes": [
				{
					"key": { "date": 1 },
					"name": "date_index"
				} 
			]
		}`)
		if err != nil {
			return errors.New(ErrPrepareDatabse)
		}

		log.Println("Mongo: \"migrations\" collection created")
	}

	return nil
}

func (driver *Mongo) GetVersion() (string, error) {
	client, err := driver.connect()
	if err != nil {
		return "", nil
	}
	defer client.Disconnect(driver.ctx)

	filter := bson.D{{
		Key: "status", Value: bson.D{{
			Key: "$in", Value: []string{
				string(migrate.MigrateFinished),
				string(migrate.MigrateRunning),
			},
		}},
	}}
	opt := options.FindOne().SetSort(bson.D{{Key: "date", Value: -1}})
	coll := client.Database(driver.database).Collection("migrations")
	result := coll.FindOne(driver.ctx, filter, opt)

	if result.Err() != nil {
		if result.Err() == mongoDb.ErrNoDocuments {
			return string(migrate.NilVersion), nil
		}
		return "", result.Err()
	}

	var migration migrate.Migration
	result.Decode(&migration)

	return migration.Version, nil
}

func (driver *Mongo) RunMigration(migr *migrate.Migration) error {
	// create a record in migration collection
	if err := driver.insertMigrationSetId(migr); err != nil {
		return err
	}
	// run migration
	_, err := driver.executeCommand(string(migr.Body))
	// update migration record
	if err != nil {
		driver.updateMigrationStatus(migr, migrate.MigrateFailed)
		return err
	} else {
		driver.updateMigrationStatus(migr, migrate.MigrateFinished)
		return nil
	}
}

func (driver *Mongo) executeCommand(jsonCmd string) (bson.A, error) {
	client, err := driver.connect()
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(driver.ctx)
	// prepare command: convert json to bson.D
	var cmd bson.D
	err = bson.UnmarshalExtJSON([]byte(jsonCmd), true, &cmd)
	if err != nil {
		return nil, errors.New(ErrInvalidJsonCommand)
	}
	// execute
	var result bson.M
	err = client.Database(driver.database).RunCommand(driver.ctx, cmd).Decode(&result)
	if err != nil {
		return nil, err
	}
	// read result
	cursor, _ := result["cursor"].(bson.M)
	firstBatch, ok := cursor["firstBatch"].(bson.A)
	if !ok {
		return nil, nil
	}

	return firstBatch, nil
}

func (driver *Mongo) connect() (*mongoDb.Client, error) {
	opts := options.Client()
	opts.SetConnectTimeout(connectionTimeOut)
	opts.SetSocketTimeout(driver.timeOut)
	opts.ApplyURI(driver.conStr)

	client, err := mongoDb.Connect(driver.ctx, opts)

	return client, err
}

func (driver *Mongo) insertMigrationSetId(migr *migrate.Migration) error {
	// prepare document
	id := primitive.NewObjectID().Hex()
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	json := fmt.Sprintf(`{
		"_id": "%v",
		"version": "%v",
		"name": "%v",
		"status": "%v",
		"date": "%v"
	}`, id, migr.Version, migr.Name, migrate.MigrateRunning, currentTime)
	var doc bson.D
	bson.UnmarshalExtJSON([]byte(json), true, &doc)
	// insert document
	client, err := driver.connect()
	if err != nil {
		return err
	}
	defer client.Disconnect(driver.ctx)
	coll := client.Database(driver.database).Collection("migrations")
	_, err = coll.InsertOne(driver.ctx, doc)
	if err != nil {
		return err
	}
	// update migration id
	migr.Id = id

	return nil
}

func (driver *Mongo) updateMigrationStatus(migr *migrate.Migration, status migrate.MigrateStatus) error {
	json := `
		{
			"update": "migrations",
			"updates": [
				{
					"q": { "_id": "%v" },
					"u": { "$set": { "status": "%v" } }
				}
			]
		}
	`
	json = fmt.Sprintf(json, migr.Id, status)
	_, err := driver.executeCommand(json)
	if err != nil {
		return err
	}
	return nil
}
