package external

import (
	"duolingo/repositories/campaign_message/user_repository/models"
)

type UserRepository interface {
	InsertManyUsers(users []*models.User) ([]*models.User, error)
	DeleteUsersByIds(ids []string) error
	DeleteUsersByCampaign(campaign string) error
	GetListUsersByIds(ids []string) ([]*models.User, error)
	GetListUsersByCampaign(campaign string) ([]*models.User, error)
	CountUserDevicesForCampaign(campaign string) (uint64, error)
}
