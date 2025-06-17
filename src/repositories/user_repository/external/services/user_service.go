package services

type UserService interface {
	CountDevicesForCampaign(campaign string) (uint64, error)
	GetDeviceTokensForCampaign(campaign string, offset uint64, limit uint64) ([]string, error)
}
