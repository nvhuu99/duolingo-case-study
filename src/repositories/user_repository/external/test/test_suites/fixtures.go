package test_suites

import "duolingo/models"

var (
	testOnlyCampaign = "testOnlyCampaign"
	testDevices      = []*models.UserDevice{
		{Token: "fake_token_1", Platform: "fake_platform"},
		{Token: "fake_token_2", Platform: "fake_platform"},
	}
)
