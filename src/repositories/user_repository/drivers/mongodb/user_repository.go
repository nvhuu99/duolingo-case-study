package mongodb

import (
	"context"

	"duolingo/models"
	driver_cmd "duolingo/repositories/user_repository/drivers/mongodb/commands"
	driver_results "duolingo/repositories/user_repository/drivers/mongodb/commands/results"
	cmd "duolingo/repositories/user_repository/external/commands"
	results "duolingo/repositories/user_repository/external/commands/results"

	connection "duolingo/libraries/connection_manager/drivers/mongodb"

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
		_, insertErr := collection.InsertMany(ctx, bsonData)
		return insertErr
	})
	return users, err
}

func (repo *UserRepo) DeleteUsersByIds(ids []string) error {
	deletion := driver_cmd.NewDeleteUsersCommand()
	deletion.SetFilterIds(ids)
	return repo.DeleteUsers(deletion)
}

func (repo *UserRepo) DeleteUsers(command cmd.DeleteUsersCommand) error {
	mongoCmd, ok := command.(*driver_cmd.DeleteUsersCommand)
	if !ok {
		panic(ErrInvalidCommandType)
	}
	if err := mongoCmd.Build(); err != nil {
		return err
	}
	var err error
	timeout := repo.GetWriteTimeout()
	err = repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Client) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		_, deleteErr := collection.DeleteMany(ctx, mongoCmd.GetFilters())
		return deleteErr
	})
	return err
}

func (repo *UserRepo) GetListUsersByIds(ids []string) ([]*models.User, error) {
	command := driver_cmd.NewListUsersCommand()
	command.SetFilterIds(ids)
	command.SetSortById(cmd.OrderASC)
	return repo.GetListUsers(command)
}

func (repo *UserRepo) GetListUsers(command cmd.ListUsersCommand) ([]*models.User, error) {
	mongoCmd, ok := command.(*driver_cmd.ListUsersCommand)
	if !ok {
		panic(ErrInvalidCommandType)
	}
	if err := mongoCmd.Build(); err != nil {
		return nil, err
	}
	var err error
	var users []*models.User
	timeout := repo.GetReadTimeout()
	err = repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Client) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		cursor, cursorErr := collection.Find(
			ctx,
			mongoCmd.GetFilters(),
			mongoCmd.GetOptions(),
		)
		if cursorErr != nil {
			return cursorErr
		}
		defer cursor.Close(ctx)
		return cursor.All(ctx, &users)
	})
	return users, err
}

func (repo *UserRepo) GetListUserDevices(command cmd.ListUserDevicesCommand) (
	[]*models.UserDevice,
	error,
) {
	mongoCmd, ok := command.(*driver_cmd.ListUserDevicesCommand)
	if !ok {
		panic(ErrInvalidCommandType)
	}
	if err := mongoCmd.Build(); err != nil {
		return nil, err
	}
	var err error
	var userDevices []*models.UserDevice
	timeout := repo.GetReadTimeout()
	err = repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Client) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		cursor, cursorErr := collection.Aggregate(ctx, mongoCmd.GetPipeline())
		if cursorErr != nil {
			return cursorErr
		}
		defer cursor.Close(ctx)
		return cursor.All(ctx, &userDevices)
	})
	return userDevices, err
}

func (repo *UserRepo) AggregateUsers(command cmd.AggregateUsersCommand) (
	results.UsersAggregationResult,
	error,
) {
	mongoCmd, ok := command.(*driver_cmd.AggregateUsersCommand)
	if !ok {
		panic(ErrInvalidCommandType)
	}
	if err := mongoCmd.Build(); err != nil {
		return nil, err
	}
	var err error
	var result = new(driver_results.UsersAggregationResult)
	timeout := repo.GetReadTimeout()
	err = repo.ExecuteClosure(timeout, func(ctx context.Context, conn *mongo.Client) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		cursor, cursorErr := collection.Aggregate(ctx, mongoCmd.GetPipeline())
		if cursorErr != nil {
			return cursorErr
		}
		defer cursor.Close(ctx)
		if cursor.Next(ctx) {
			if decodeErr := cursor.Decode(result); decodeErr != nil {
				return decodeErr
			}
		}
		return nil
	})
	return result, err
}
