package commands

type SortOrder string

const (
	OrderASC  SortOrder = "asc"
	OrderDESC SortOrder = "desc"
)

type ListUsersCommand interface {
	SetFilterIds(ids []string)
	SetFilterCampaign(campaign string)
	SetFilterOnlyEmailVerified()

	SetPagination(offset uint64, limit uint64)
	SetSortById(order SortOrder)

	Build() error
}
