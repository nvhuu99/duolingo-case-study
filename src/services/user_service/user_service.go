package user_service

import (
	"context"
	"duolingo/models"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/repositories/user_repository/external/commands"
	"errors"
)

type UserService struct {
	user_repo.UserRepository
	user_repo.UserRepoFactory
}

func NewUserService(
	factory user_repo.UserRepoFactory,
	repo user_repo.UserRepository,
) *UserService {
	return &UserService{
		UserRepoFactory: factory,
		UserRepository:  repo,
	}
}

func (service *UserService) CountDevicesForCampaign(ctx context.Context, campaign string) (int64, error) {
	cmd := service.MakeAggregateUsersCommand()
	cmd.SetFilterCampaign(campaign)
	cmd.SetFilterOnlyEmailVerified()
	cmd.AddAggregationSumUserDevices()

	result, err := service.AggregateUsers(ctx, cmd)
	if err != nil {
		return 0, errors.New("failed to get devices count for campaign")
	}
	return result.GetCountUserDevices(), nil
}

func (service *UserService) GetDevicesForCampaign(
	ctx context.Context,
	campaign string,
	offset int64,
	limit int64,
) ([]*models.UserDevice, error) {
	query := service.MakeListUserDevicesCommand()
	query.SetFilterCampaign(campaign)
	query.SetFilterOnlyEmailVerified()
	query.SetPagination(offset, limit)
	query.SetSortById(commands.OrderASC)

	devices, err := service.GetListUserDevices(ctx, query)
	if err != nil {
		return nil, err
	}
	return devices, nil
}
