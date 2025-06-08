package mongodb

import (
	"context"
	connection "duolingo/libraries/connection_manager/drivers/mongodb"
	cmd "duolingo/repositories/campaign_message/user_repository/drivers/mongodb/command_builders"
	"duolingo/repositories/campaign_message/user_repository/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	connection.MongoClient

	databaseName   string
	collectionName string
}

func NewUserRepo(
	client *connection.MongoClient,
	databaseName string,
	collectionName string,
) *UserRepo {
	return &UserRepo{
		MongoClient:    *client,
		databaseName:   databaseName,
		collectionName: collectionName,
	}
}

func (repo *UserRepo) InsertManyUsers(users []*models.User) ([]*models.User, error) {
	bsonData := make([]any, len(users))
	for i := range users {
		if users[i].Id == "" {
			users[i].Id = uuid.NewString()
		}
		bsonData[i] = users[i]
	}
	var err error
	timeout := repo.GetWriteTimeout()
	err = repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Client) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		_, err = collection.InsertMany(ctx, bsonData)
		return err
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (repo *UserRepo) DeleteUsersByIds(ids []string) error {
	builder := new(cmd.UserDeleteCommandBuilder)
	builder.WithUserIdsFilter(ids)
	return repo.deleteUsers(builder)
}

func (repo *UserRepo) DeleteUsersByCampaign(campaign string) error {
	builder := new(cmd.UserDeleteCommandBuilder)
	builder.WithCampaignFilter(campaign)
	return repo.deleteUsers(builder)
}

func (repo *UserRepo) GetListUsersByIds(ids []string) ([]*models.User, error) {
	builder := new(cmd.UserListCommandBuilder)
	builder.WithUserIdsFilter(ids)
	return repo.getListUsers(builder)
}

func (repo *UserRepo) GetListUsersByCampaign(campaign string) ([]*models.User, error) {
	builder := new(cmd.UserListCommandBuilder)
	builder.WithCampaignFilter(campaign)
	return repo.getListUsers(builder)
}

func (repo *UserRepo) CountUserDevicesForCampaign(campaign string) (uint64, error) {
	builder := new(cmd.UserDevicesAggregationCommandBuilder)
	builder.WithCampaignFilter(campaign)
	builder.WithEmailVerifiedOnlyFilter()
	builder.WithSumUserDevicesAggregation()
	if err := builder.Build(); err != nil {
		return 0, err
	}
	var err error
	var total uint64
	timeout := repo.GetReadTimeout()
	err = repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Client) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		cursor, cursorErr := collection.Aggregate(ctx, builder.GetPipeline())
		if cursorErr != nil {
			return cursorErr
		}
		defer cursor.Close(ctx)
		if cursor.Next(ctx) {
			var result struct {
				Total uint64 `bson:"total"`
			}
			if decodeErr := cursor.Decode(&result); decodeErr != nil {
				return decodeErr
			}
			total = result.Total
		}
		return nil
	})
	return total, err
}

func (repo *UserRepo) getListUsers(builder *cmd.UserListCommandBuilder) ([]*models.User, error) {
	var err error
	var users []*models.User
	timeout := repo.GetReadTimeout()
	err = repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Client) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		if err = builder.Build(); err != nil {
			return err
		}
		cursor, cursorErr := collection.Find(ctx, builder.GetFilters(), builder.GetOptions())
		if cursorErr != nil {
			return cursorErr
		}
		defer cursor.Close(ctx)
		return cursor.All(ctx, &users)
	})
	return users, err
}

func (repo *UserRepo) deleteUsers(builder *cmd.UserDeleteCommandBuilder) error {
	var err error
	timeout := repo.GetWriteTimeout()
	repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Client) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		if err = builder.Build(); err != nil {
			return err
		}
		_, err = collection.DeleteMany(ctx, builder.GetFilters())
		return err
	})
	return err
}
