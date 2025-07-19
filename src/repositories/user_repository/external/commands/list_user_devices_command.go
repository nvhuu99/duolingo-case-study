package commands

type ListUserDevicesCommand interface {
	SetFilterIds(ids []string)
	SetFilterCampaign(campaign string)
	SetFilterOnlyEmailVerified()

	SetPagination(offset int64, limit int64)
	SetSortById(order SortOrder)

	Build() error
}
