package commands

type ListUserDevicesCommand interface {
	SetFilterIds(ids []string)
	SetFilterCampaign(campaign string)
	SetFilterOnlyEmailVerified()

	SetPagination(offset uint64, limit uint64)
	SetSortById(order SortOrder)

	Build() error
}
