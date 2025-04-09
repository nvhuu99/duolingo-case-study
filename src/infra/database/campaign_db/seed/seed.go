package main

import (
	"context"
	config "duolingo/lib/config_reader/driver/reader/json"
	"duolingo/model"
	campaigndb "duolingo/repository/campaign_db"
	"log"
	"path/filepath"

	"github.com/google/uuid"
)

func main() {
	conf := config.NewJsonReader(filepath.Join(".", "config"))

	userRepo := campaigndb.NewUserRepo(
		context.Background(),
		conf.Get("db.campaign.name", ""),
	)
	userRepo.SetConnection(
		conf.Get("db.campaign.host", ""),
		conf.Get("db.campaign.port", ""),
		conf.Get("db.campaign.user", ""),
		conf.Get("db.campaign.password", ""),
	)

	users := make([]*model.CampaignUser, 100)
	for i := 0; i < 50; i++ {
		users[i] = &model.CampaignUser{
			Campaign:       "superbowl",
			LastName:       "John",
			FirstName:      "Doe",
			DeviceToken:    uuid.New().String(),
			NativeLanguage: "EN",
			Membership:     model.Premium,
			SortValue:      "1",
		}
	}

	for i := 50; i < 100; i++ {
		users[i] = &model.CampaignUser{
			Campaign:       "superbowl",
			LastName:       "John",
			FirstName:      "Doe",
			DeviceToken:    uuid.New().String(),
			NativeLanguage: "EN",
			Membership:     model.FreeTier,
			SortValue:      "2",
		}
	}

	_, err := userRepo.InsertUsers(users)
	if err != nil {
		log.Fatalln(err)
	}
	
	log.Println("seeder run successfully")
}
