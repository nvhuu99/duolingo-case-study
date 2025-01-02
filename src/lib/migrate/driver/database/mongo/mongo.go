package mongo

import (
	"context"
	migrate "duolingo/lib/migrate"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	"net/url"

	"go.mongodb.org/mongo-driver/bson"
	mongoDb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectionTimeOut = 10 * time.Second
	defaultTimeOut    = 30 * time.Second
)

var (
	errInvalidJsonCommand = "Mongo: failed to parse json command"
)

type Mongo struct {
	ctx      context.Context
	uri   string
	database string
	batchNumber int
	timeOut time.Duration
}

func New(ctx context.Context) *Mongo {
	driver := Mongo{}
	driver.ctx = ctx
	driver.timeOut = defaultTimeOut

	return &driver
}

func (driver *Mongo) SetDatabase(database string) {
	driver.database = database
}

func (driver *Mongo) GetFileExt() string {
	return ".json"
}

func (driver *Mongo) SetConnection(host string, port string, usr string, pwd string) {
	driver.uri = fmt.Sprintf(
		"mongodb://%v:%v@%v:%v/", 
		url.QueryEscape(usr), 
		url.QueryEscape(pwd),
		host,
		port, 
	)
}

func (driver *Mongo) SetOperationTimeOut(duration time.Duration) {
	driver.timeOut = duration
}

// PrepareDatabase initializes the "migrations" collection and retrieves the last batch number
func (driver *Mongo) PrepareDatabase() error {
	// Check if the "migrations" collection exists
	listResult, err := driver.executeCommand(`{
		"listCollections": 1,
		"filter": { "name": "migrations" },
		"nameOnly": true
	}`)
	if err != nil {
		return err
	}
	// Create the "migrations" collection if it doesn't exist
	if len(listResult) == 0 {
		createSchema := `{
			"create": "migrations",
			"validator": {
				"$jsonSchema": {
					"bsonType": "object",
					"required": [
						"id",
						"name",
						"batchNumber",
						"status"
					],
					"properties": {
						"status": {
							"bsonType": "string",
							"enum": ["finished", "failed"]
						}
					}
				}
			},
			"validationAction": "error",
			"validationLevel": "strict"
		}`
		if _, err := driver.executeCommand(createSchema); err != nil {
			return err
		}

		log.Println("Mongo: \"migrations\" collection created")
	}
	// Retrieve the last batch number from the "migrations" collection
	var lastBatchNum int
	client, err := driver.connect()
	defer client.Disconnect(driver.ctx)
	if err != nil {
		return err
	}
	filter := bson.D{{
		Key: "status", Value: string(migrate.MigrateFinished),
	}}
	opt := options.FindOne().SetSort(bson.D{
		{Key: "batchNumber", Value: -1},
		{Key: "id", Value: -1},
	})
	coll := client.Database(driver.database).Collection("migrations")
	result := coll.FindOne(driver.ctx, filter, opt)
	if result.Err() != nil {
		if result.Err() == mongoDb.ErrNilDocument ||
			result.Err() == mongoDb.ErrNoDocuments {
			lastBatchNum = 0
		} else {
			return result.Err()
		}
	} else {
		var migr bson.M
		result.Decode(&migr)
		lastBatchNum, _ = strconv.Atoi(migr["batchNumber"].(string))
	}

	// Set the next batch number.
	driver.batchNumber = lastBatchNum + 1

	return nil
}

func (driver *Mongo) BatchNumber() int {
	return driver.batchNumber
}

// Retrieves the last batch of migrations from the "migrations" collection.
func (driver *Mongo) LastBatch() ([]migrate.Migration, error) {
	client, err := driver.connect()
	if err != nil {
		return []migrate.Migration{}, err
	}
	filter := bson.D{{
		Key: "batchNumber", Value: strconv.Itoa(driver.batchNumber - 1),
	}}
	opt := options.Find().SetSort(bson.D{
		{Key: "id", Value: 1},
	})
	coll := client.Database(driver.database).Collection("migrations")
	result, err := coll.Find(driver.ctx, filter, opt)
	client.Disconnect(driver.ctx)
	if err != nil {
		return []migrate.Migration{}, err
	}

	var records []bson.M
	err = result.All(driver.ctx, &records)
	if err != nil {
		return []migrate.Migration{}, nil
	}

	migrations := make([]migrate.Migration, len(records))
	for i, item := range records {
		migrations[i] = migrate.Migration{ 
			Id: item["id"].(string),
			Name: item["name"].(string),
			Status: migrate.MigrateStatus(item["status"].(string)),
			BatchNumber: item["batchNumber"].(string),
		}
	}

	return migrations, nil
}

func (driver *Mongo) RunMigration(migr *migrate.Migration) error {
	_, err := driver.executeCommand(string(migr.Body))
	if err != nil {
		return err
	}

	return nil
}

func (driver *Mongo) SaveMigrationRecord(migr *migrate.Migration) error {
	// prepare document
	id := strconv.FormatInt(time.Now().UnixMicro(), 10)
	json := fmt.Sprintf(`{
		"id": "%v",
		"name": "%v",
		"batchNumber": "%v",
		"status": "%v"
	}`, id, migr.Name, migr.BatchNumber, migr.Status)
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

	return nil
}

func (driver *Mongo) DeleteMigrationRecord(migr *migrate.Migration) error {
	cmd := `{
		"delete": "migrations",
		"deletes": [ { "q": { "id": "%v"}, "limit": 1 } ]
	}`
	cmd = fmt.Sprintf(cmd, migr.Id)
	_, err := driver.executeCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// Runs a JSON-formatted MongoDB command.
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
		return nil, errors.New(errInvalidJsonCommand)
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
	opts.ApplyURI(driver.uri)

	client, err := mongoDb.Connect(driver.ctx, opts)

	return client, err
}
