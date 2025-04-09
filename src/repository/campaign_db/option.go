package campaigndb

import "duolingo/model"

type ListUserOptions struct {
	Skip     int
	Limit    int
	Campaign string

	CursorMode bool
	CursorFunc func(usr *model.CampaignUser) bool
}
