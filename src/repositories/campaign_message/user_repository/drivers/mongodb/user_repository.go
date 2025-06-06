package mongodb

import (
	"context"
	cmd "duolingo/repositories/campaign_message/user_repository/drivers/mongodb/command_builders"
	"duolingo/repositories/campaign_message/user_repository/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	client *Client
}

func (repo *UserRepo) InsertManyUsers(users []*models.User) ([]*models.User, error) {
	bsonData := make([]any, len(users))
	for i := range users {
		users[i].Id = uuid.NewString()
		bsonData[i] = users[i]
	}

	var err error
	client := repo.client
	timeout := client.GetWriteTimeout()
	client.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Collection) error {
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
	var total uint64
	client := repo.client
	timeout := client.GetReadTimeout()
	err := client.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Collection) error {
		builder := new(cmd.UserDevicesAggregationCommandBuilder)
		builder.WithCampaignFilter(campaign)
		builder.WithEmailVerifiedOnlyFilter()
		builder.WithSumUserDevicesAggregation()
		if err := builder.Build(); err != nil {
			return err
		}

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
	client := repo.client
	timeout := client.GetReadTimeout()
	err := client.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Collection) error {
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
	client := repo.client
	timeout := client.GetWriteTimeout()
	return client.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Collection) error {
		if err := builder.Build(); err != nil {
			return err
		}
		_, err := conn.DeleteMany(ctx, builder.GetFilters())
		return err
	})
}
