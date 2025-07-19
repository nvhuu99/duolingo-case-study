package data

import (
	"duolingo/models"
	"time"
)

var (
	TestCampaignPrimary   = "test_campaign_primary"
	TestCampaignSecondary = "test_campaign_secondary"

	TestCampaignUserIdsMap = map[string][]string{
		TestCampaignPrimary: {
			"user_1",
			"user_2",
			"user_3",
			"user_4",
			"user_5",
		},
		TestCampaignSecondary: {
			"user_1",
			"user_2",
		},
	}

	TestPlatforms = []string{
		"android",
		"ios",
	}

	TestDevices = []*models.UserDevice{
		{Token: "user_1_device_1", Platform: "android"},
		{Token: "user_1_device_2", Platform: "ios"},
		{Token: "user_2_device_1", Platform: "android"},
		{Token: "user_2_device_2", Platform: "ios"},
		{Token: "user_3_device_1", Platform: "android"},
		{Token: "user_3_device_2", Platform: "ios"},
		{Token: "user_4_device_1", Platform: "android"},
		{Token: "user_4_device_2", Platform: "ios"},
		{Token: "user_5_device_1", Platform: "android"},
		{Token: "user_5_device_2", Platform: "ios"},
	}

	TestUserIds = []string{
		"user_1",
		"user_2",
		"user_3",
		"user_4",
		"user_5",
	}

	TestUsersEmailUnverifiedIds = []string{
		"user_6",
		"user_7",
	}

	TestUsers = []*models.User{
		{
			Id:        "user_1",
			Campaigns: []string{TestCampaignPrimary, TestCampaignSecondary},
			Devices: []*models.UserDevice{
				{Token: "user_1_device_1", Platform: "android"},
				{Token: "user_1_device_2", Platform: "ios"},
			},
			EmailVerifiedAt: time.Now().UTC().Add(-1 * time.Hour),
		},
		{
			Id:        "user_2",
			Campaigns: []string{TestCampaignPrimary, TestCampaignSecondary},
			Devices: []*models.UserDevice{
				{Token: "user_2_device_1", Platform: "android"},
				{Token: "user_2_device_2", Platform: "ios"},
			},
			EmailVerifiedAt: time.Now().UTC().Add(-1 * time.Hour),
		},
		{
			Id:        "user_3",
			Campaigns: []string{TestCampaignPrimary},
			Devices: []*models.UserDevice{
				{Token: "user_3_device_1", Platform: "android"},
				{Token: "user_3_device_2", Platform: "ios"},
			},
			EmailVerifiedAt: time.Now().UTC().Add(-1 * time.Hour),
		},
		{
			Id:        "user_4",
			Campaigns: []string{TestCampaignPrimary},
			Devices: []*models.UserDevice{
				{Token: "user_4_device_1", Platform: "android"},
				{Token: "user_4_device_2", Platform: "ios"},
			},
			EmailVerifiedAt: time.Now().UTC().Add(-1 * time.Hour),
		},
		{
			Id:        "user_5",
			Campaigns: []string{TestCampaignPrimary},
			Devices: []*models.UserDevice{
				{Token: "user_5_device_1", Platform: "android"},
				{Token: "user_5_device_2", Platform: "ios"},
			},
			EmailVerifiedAt: time.Now().UTC().Add(-1 * time.Hour),
		},
	}

	TestUsersEmailUnverified = []*models.User{
		{Id: "user_6"},
		{Id: "user_7"},
	}
)
