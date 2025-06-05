package user_repository

import "duolingo/repositories/campaign_message/user_repository/models"

type UserRepository interface {
	InsertManyUsers(rows []*models.User) ([]*models.User, error)
	DeleteUsers(users []*models.User) error
	DeleteUsersByIds(ids []string) error
	GetListUsersByIds(ids []string) ([]*models.User, error)
	GetListUsersByCampaign(campaign string) ([]*models.User, error)
	CountUserDevicesForCampaign(campaign string) (uint64, error)
}
