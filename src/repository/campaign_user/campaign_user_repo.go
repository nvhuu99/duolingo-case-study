package campaign_user

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectionTimeOut = 10 * time.Second
	defaultTimeOut    = 30 * time.Second
)

type UserRepo struct {
	uri      string
	database string
	client   *mongo.Client
	ctx      context.Context
}

func NewUserRepo(ctx context.Context, database string) *UserRepo {
	repo := UserRepo{}
	repo.ctx = ctx
	repo.database = database

	return &repo
}

func (repo *UserRepo) SetConnection(host string, port string, usr string, pwd string) error {
	repo.uri = fmt.Sprintf(
		"mongodb://%v:%v@%v:%v/",
		url.QueryEscape(usr),
		url.QueryEscape(pwd),
		host,
		port,
	)

	opts := options.Client()
	opts.SetConnectTimeout(connectionTimeOut)
	opts.SetSocketTimeout(defaultTimeOut)
	opts.ApplyURI(repo.uri)

	client, err := mongo.Connect(repo.ctx, opts)
	if err != nil {
		return err
	}
	repo.client = client

	return nil
}

func (repo *UserRepo) CountCampaignMsgReceivers(campaign string, timestamp time.Time) (int, error) {
	filter := repo.campaignMsgReceiversQuery(campaign, timestamp)
	coll := repo.client.Database(repo.database).Collection("campaign_users")
	count, err := coll.CountDocuments(repo.ctx, filter)
	return int(count), err
}

func (repo *UserRepo) ListCampaignMsgReceiverTokens(campaign string, timestamp time.Time, opts *QueryOptions) ([]string, error) {
	filter := repo.campaignMsgReceiversQuery(campaign, timestamp)
	opt := options.Find()
	if opts.Skip >= 0 {
		opt.SetSkip(opts.Skip)
	}
	if opts.Limit > 0 {
		opt.SetLimit(opts.Limit)
	}
	collection := repo.client.Database(repo.database).Collection("campaign_users")

	cursor, err := collection.Find(repo.ctx, filter, opt)
	if err != nil {
		return []string{}, err
	}
	defer cursor.Close(repo.ctx)

	tokens := []string{}
	for cursor.Next(repo.ctx) {
		user := new(CampaignUser)
		if err := cursor.Decode(user); err != nil {
			return []string{}, err
		}
		tokens = append(tokens, user.DeviceToken)
	}

	if err := cursor.Err(); err != nil {
		return []string{}, err
	}

	return tokens, nil
}

func (repo *UserRepo) InsertUsers(users []*CampaignUser) ([]any, error) {
	bsonData := make([]interface{}, len(users))
	for i, usr := range users {
		bsonData[i] = usr
	}

	collection := repo.client.Database(repo.database).Collection("campaign_users")
	result, err := collection.InsertMany(repo.ctx, bsonData)
	if err != nil {
		return []any{}, err
	}

	return result.InsertedIDs, nil
}

func (repo *UserRepo) campaignMsgReceiversQuery(campaign string, timestamp time.Time) bson.M {
	primitiveTime := primitive.NewDateTimeFromTime(timestamp)
	return bson.M{
		"campaign":    campaign,
		"verified_at": bson.M{"$lte": primitiveTime},
	}
}
