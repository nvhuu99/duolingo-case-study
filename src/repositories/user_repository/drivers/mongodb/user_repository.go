package mongodb

import (
	"context"

	"duolingo/models"
	driver_cmd "duolingo/repositories/user_repository/drivers/mongodb/commands"
	driver_results "duolingo/repositories/user_repository/drivers/mongodb/commands/results"
	cmd "duolingo/repositories/user_repository/external/commands"
	results "duolingo/repositories/user_repository/external/commands/results"

	connection "duolingo/libraries/connection_manager/drivers/mongodb"
	events "duolingo/libraries/events/facade"

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

func (repo *UserRepo) InsertManyUsers(ctx context.Context, users []*models.User) ([]*models.User, error) {
	var err error

	evt := events.Start(ctx, "user_repo.insert_many_users", nil)
	defer events.End(evt, true, err, nil)

	bsonData := make([]any, len(users))
	for i := range users {
		if users[i].Id == "" {
			users[i].Id = uuid.NewString()
		}
		bsonData[i] = users[i]
	}

	timeout := repo.GetWriteTimeout()
	err = repo.ExecuteClosure(evt.Context(), timeout, func(
		timeoutCtx context.Context,
		conn *mongo.Client,
	) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		_, insertErr := collection.InsertMany(timeoutCtx, bsonData)
		return insertErr
	})

	return users, err
}

func (repo *UserRepo) DeleteUsersByIds(ctx context.Context, ids []string) error {
	var err error

	evt := events.Start(ctx, "user_repo.delete_by_user_id", nil)
	defer events.End(evt, true, err, nil)

	deletion := driver_cmd.NewDeleteUsersCommand()
	deletion.SetFilterIds(ids)
	err = repo.DeleteUsers(evt.Context(), deletion)

	return err
}

func (repo *UserRepo) DeleteUsers(ctx context.Context, command cmd.DeleteUsersCommand) error {
	var err error

	evt := events.Start(ctx, "user_repo.delete_by_users", nil)
	defer events.End(evt, true, err, nil)

	mongoCmd, ok := command.(*driver_cmd.DeleteUsersCommand)
	if !ok {
		panic(ErrInvalidCommandType)
	}
	if err := mongoCmd.Build(); err != nil {
		return err
	}

	timeout := repo.GetWriteTimeout()
	err = repo.ExecuteClosure(evt.Context(), timeout, func(
		timeoutCtx context.Context,
		conn *mongo.Client,
	) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		_, deleteErr := collection.DeleteMany(ctx, mongoCmd.GetFilters())
		return deleteErr
	})

	return err
}

func (repo *UserRepo) GetListUsersByIds(ctx context.Context, ids []string) ([]*models.User, error) {
	var err error
	var users []*models.User

	evt := events.Start(ctx, "user_repo.get_list_users_by_ids", nil)
	defer events.End(evt, true, err, nil)

	command := driver_cmd.NewListUsersCommand()
	command.SetFilterIds(ids)
	command.SetSortById(cmd.OrderASC)
	users, err = repo.GetListUsers(ctx, command)

	return users, err
}

func (repo *UserRepo) GetListUsers(
	ctx context.Context,
	command cmd.ListUsersCommand,
) ([]*models.User, error) {
	var err error
	var users []*models.User

	evt := events.Start(ctx, "user_repo.get_list_users", nil)
	defer events.End(evt, true, err, nil)

	mongoCmd, ok := command.(*driver_cmd.ListUsersCommand)
	if !ok {
		panic(ErrInvalidCommandType)
	}
	if err = mongoCmd.Build(); err != nil {
		return nil, err
	}

	timeout := repo.GetReadTimeout()
	err = repo.ExecuteClosure(evt.Context(), timeout, func(
		timeoutCtx context.Context,
		conn *mongo.Client,
	) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		cursor, cursorErr := collection.Find(
			timeoutCtx,
			mongoCmd.GetFilters(),
			mongoCmd.GetOptions(),
		)
		if cursorErr != nil {
			return cursorErr
		}
		defer cursor.Close(timeoutCtx)
		return cursor.All(timeoutCtx, &users)
	})

	return users, err
}

func (repo *UserRepo) GetListUserDevices(ctx context.Context, command cmd.ListUserDevicesCommand) (
	[]*models.UserDevice,
	error,
) {
	var err error
	var userDevices []*models.UserDevice

	evt := events.Start(ctx, "user_repo.get_list_user_devices", nil)
	defer events.End(evt, true, err, nil)

	mongoCmd, ok := command.(*driver_cmd.ListUserDevicesCommand)
	if !ok {
		panic(ErrInvalidCommandType)
	}
	if err := mongoCmd.Build(); err != nil {
		return nil, err
	}

	timeout := repo.GetReadTimeout()
	err = repo.ExecuteClosure(evt.Context(), timeout, func(
		timeoutCtx context.Context,
		conn *mongo.Client,
	) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		cursor, cursorErr := collection.Aggregate(timeoutCtx, mongoCmd.GetPipeline())
		if cursorErr != nil {
			return cursorErr
		}
		defer cursor.Close(timeoutCtx)
		return cursor.All(timeoutCtx, &userDevices)
	})

	return userDevices, err
}

func (repo *UserRepo) AggregateUsers(ctx context.Context, command cmd.AggregateUsersCommand) (
	results.UsersAggregationResult,
	error,
) {
	var err error
	var result = new(driver_results.UsersAggregationResult)

	evt := events.Start(ctx, "user_repo.aggregate_users", nil)
	defer events.End(evt, true, err, nil)

	mongoCmd, ok := command.(*driver_cmd.AggregateUsersCommand)
	if !ok {
		panic(ErrInvalidCommandType)
	}
	if err := mongoCmd.Build(); err != nil {
		return nil, err
	}

	timeout := repo.GetReadTimeout()
	err = repo.ExecuteClosure(evt.Context(), timeout, func(
		timeoutCtx context.Context,
		conn *mongo.Client,
	) error {
		collection := conn.Database(repo.databaseName).Collection(repo.collectionName)
		cursor, cursorErr := collection.Aggregate(timeoutCtx, mongoCmd.GetPipeline())
		if cursorErr != nil {
			return cursorErr
		}
		defer cursor.Close(timeoutCtx)
		if cursor.Next(timeoutCtx) {
			if decodeErr := cursor.Decode(result); decodeErr != nil {
				return decodeErr
			}
		}
		return nil
	})

	return result, err
}
