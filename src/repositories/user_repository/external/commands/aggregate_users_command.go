package commands

type AggregateUsersCommand interface {
	SetFilterIds(ids []string)
	SetFilterCampaign(campaign string)
	SetFilterOnlyEmailVerified()
	AddAggregationSumUserDevices()
	Build() error
}
