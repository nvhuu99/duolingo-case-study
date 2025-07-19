package commands

type DeleteUsersCommand interface {
	SetFilterIds(ids []string)
	SetFilterCampaign(campaign string)
	Build() error
}
