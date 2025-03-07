package campaigndb

import (
	"context"
	"duolingo/model"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectionTimeOut = 10 * time.Second
	defaultTimeOut    = 30 * time.Second
)

type UserRepo struct {
	uri			string
	database	string
	ctx			context.Context
}

func NewUserRepo(ctx context.Context, database string) *UserRepo {
	repo := UserRepo{}
	repo.ctx = ctx
	repo.database = database

	return &repo
}

func (repo *UserRepo) SetConnection(host string, port string, usr string, pwd string) {
	repo.uri = fmt.Sprintf(
		"mongodb://%v:%v@%v:%v/",
		url.QueryEscape(usr),
		url.QueryEscape(pwd),
		host,
		port,
	)
}

func (repo *UserRepo) CountUsers(campaign string) (int, error) {
	client, err := repo.connect()
	defer client.Disconnect(repo.ctx)
	if err != nil {
		return 0, err
	}

	filter := bson.D{
		{ Key: "campaign", Value: campaign },
	}
	count, err := client.Database(repo.database).
		Collection("campaign_users").
		CountDocuments(repo.ctx, filter)

	return int(count), err
}

func (repo *UserRepo) UsersList(args *ListUserOptions) ([]*model.CampaignUser, error) {
	client, err := repo.connect()
	defer client.Disconnect(repo.ctx)
	if err != nil {
		return []*model.CampaignUser{}, err
	}

	filter := bson.D {}
	if args.Campaign != "" {
		filter = append(filter, bson.E{ Key: "campaign", Value: args.Campaign })
	}

	opt := options.Find()
	if args.Skip >= 0 {
		opt.SetSkip(int64(args.Skip))
	}
	if args.Limit > 0 {
		opt.SetLimit(int64(args.Limit))
	}

	collection := client.Database(repo.database).Collection("campaign_users")
	cursor, err := collection.Find(repo.ctx, filter, opt)
	if err != nil {
		return []*model.CampaignUser{}, err
	}
	defer cursor.Close(repo.ctx)

	var users []*model.CampaignUser
	for cursor.Next(repo.ctx) {
		var user model.CampaignUser
		if err := cursor.Decode(&user); err != nil {
			return []*model.CampaignUser{}, err
		}
		if !args.CursorMode {
			users = append(users, &user)
		} else {
			flag := args.CursorFunc(&user)
			if !flag {
				return []*model.CampaignUser{}, nil
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return []*model.CampaignUser{}, err
	}

	return users, nil
}

func (repo *UserRepo) InsertUsers(users []*model.CampaignUser) ([]any, error) {
	client, err := repo.connect()
	defer client.Disconnect(repo.ctx)
	if err != nil {
		return []any{}, err
	}

	bsonData := make([]interface{}, len(users)) // Use interface{} for MongoDB compatibility
	for i, usr := range users {
		bsonData[i] = usr // Directly assign the struct
	}

	collection := client.Database(repo.database).Collection("campaign_users")
	result, err := collection.InsertMany(repo.ctx, bsonData)
	if err != nil {
		return []any{}, err
	}

	return result.InsertedIDs, nil
}

func (repo *UserRepo) connect() (*mongo.Client, error) {
	opts := options.Client()
	opts.SetConnectTimeout(connectionTimeOut)
	opts.SetSocketTimeout(defaultTimeOut)
	opts.ApplyURI(repo.uri)

	client, err := mongo.Connect(repo.ctx, opts)

	return client, err
}
