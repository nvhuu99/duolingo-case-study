package user_service

import (
	"context"
	events "duolingo/libraries/events/facade"
	"duolingo/models"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/repositories/user_repository/external/commands"
	"duolingo/repositories/user_repository/external/commands/results"
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
	var aggregateResult results.UsersAggregationResult
	var err error

	evt := events.Start(ctx, "user_service.count_devices_for_campaign", nil)
	defer events.End(evt, true, err, nil)

	cmd := service.MakeAggregateUsersCommand()
	cmd.SetFilterCampaign(campaign)
	cmd.SetFilterOnlyEmailVerified()
	cmd.AddAggregationSumUserDevices()

	aggregateResult, err = service.AggregateUsers(evt.Context(), cmd)
	if err != nil {
		return 0, err
	}

	return aggregateResult.GetCountUserDevices(), nil
}

func (service *UserService) GetDevicesForCampaign(
	ctx context.Context,
	campaign string,
	offset int64,
	limit int64,
) ([]*models.UserDevice, error) {
	var devices []*models.UserDevice
	var err error

	evt := events.Start(ctx, "user_service.get_devices_for_campaign", nil)
	defer events.End(evt, true, err, nil)

	query := service.MakeListUserDevicesCommand()
	query.SetFilterCampaign(campaign)
	query.SetFilterOnlyEmailVerified()
	query.SetPagination(offset, limit)
	query.SetSortById(commands.OrderASC)

	devices, err = service.GetListUserDevices(ctx, query)

	return devices, err
}
