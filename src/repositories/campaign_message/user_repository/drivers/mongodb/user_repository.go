package mongodb

import (
	"context"
	"duolingo/libraries/mongo_connect"
	cmd "duolingo/repositories/campaign_message/user_repository/drivers/mongodb/command_builders"
	"duolingo/repositories/campaign_message/user_repository/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	mongo_connect.Client
}

func (repo *UserRepo) InsertManyUsers(users []*models.User) ([]*models.User, error) {
	bsonData := make([]any, len(users))
	for i := range users {
		users[i].Id = uuid.NewString()
		bsonData[i] = users[i]
	}
	var err error
	timeout := repo.GetWriteTimeout()
	repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Collection) error {
		_, err = conn.InsertMany(ctx, bsonData)
		return err
	})
	return users, err
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
	var total uint64
	timeout := repo.GetReadTimeout()
	err := repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Collection) error {
		cursor, err := conn.Aggregate(ctx, builder.GetPipeline())
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		if cursor.Next(ctx) {
			var result struct {
				Total uint64 `bson:"total"`
			}
			if err := cursor.Decode(&result); err != nil {
				return err
			}
			total = result.Total
		}
		return nil
	})
	return total, err
}

func (repo *UserRepo) getListUsers(builder *cmd.UserListCommandBuilder) ([]*models.User, error) {
	var users []*models.User
	timeout := repo.GetReadTimeout()
	err := repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Collection) error {
		if err := builder.Build(); err != nil {
			return err
		}
		cursor, err := conn.Find(ctx, builder.GetFilters(), builder.GetOptions())
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		return cursor.All(ctx, &users)
	})
	return users, err
}

func (repo *UserRepo) deleteUsers(builder *cmd.UserDeleteCommandBuilder) error {
	timeout := repo.GetWriteTimeout()
	return repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Collection) error {
		if err := builder.Build(); err != nil {
			return err
		}
		_, err := conn.DeleteMany(ctx, builder.GetFilters())
		return err
	})
}
