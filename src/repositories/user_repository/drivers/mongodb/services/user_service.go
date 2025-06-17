package services

import (
	user_repo "duolingo/repositories/user_repository/external"
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

func (service *UserService) GetDeviceTokensForCampaign(
	campaign string,
	offset uint64,
	limit uint64,
) ([]string, error) {
	query := service.MakeListUsersCommand()
	query.SetFilterCampaign(campaign)
	query.SetFilterOnlyEmailVerified()
	query.SetPagination(offset, limit)

	users, err := service.GetListUsers(query)
	if err != nil {
		return []string{}, err
	}

	tokens := []string{}
	for i := range len(users) {
		tokens = append(tokens, users[i].DeviceTokens...)
	}

	return tokens, nil
}
