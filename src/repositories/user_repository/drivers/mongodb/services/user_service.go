package services

import (
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

func (service *UserService) CountDevicesForCampaign(campaign string) (uint64, error) {
	cmd := service.MakeAggregateUsersCommand()
	cmd.SetFilterCampaign(campaign)
	cmd.SetFilterOnlyEmailVerified()
	cmd.AddAggregationSumUserDevices()

	result, err := service.AggregateUsers(cmd)
	if err != nil {
		return 0, errors.New("failed to get devices count for campaign")
	}
	return result.GetCountUserDevices(), nil
}

func (service *UserService) GetDevicesForCampaign(
	campaign string,
	offset uint64,
	limit uint64,
) ([]*models.UserDevice, error) {
	query := service.MakeListUserDevicesCommand()
	query.SetFilterCampaign(campaign)
	query.SetFilterOnlyEmailVerified()
	query.SetPagination(offset, limit)
	query.SetSortById(commands.OrderASC)

	devices, err := service.GetListUserDevices(query)
	if err != nil {
		return nil, err
	}
	return devices, nil
}
