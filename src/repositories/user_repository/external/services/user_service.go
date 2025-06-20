package services

import "duolingo/models"

type UserService interface {
	CountDevicesForCampaign(campaign string) (uint64, error)
	GetDevicesForCampaign(campaign string, offset uint64, limit uint64) ([]*models.UserDevice, error)
}
