package main

import (
	"context"
	"duolingo/lib/config-reader"
	"duolingo/model"
	campaigndb "duolingo/repository/campaign-db"
	"log"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
)

func main() {
	_, caller, _, _ := runtime.Caller(0)
	dir := filepath.Dir(caller)
	conf := config.NewJsonReader(filepath.Join(dir, "..", "..", "..", "config"))

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
		users[i] = &model.CampaignUser {
			Campaign: "superbowl",
			LastName: "John",
			FirstName: "Doe",
			DeviceToken: uuid.New().String(),
			NativeLanguage: "EN",
			Membership: model.Premium,
			SortValue: "1",
		}
	}

	for i := 50; i < 100; i++ {
		users[i] = &model.CampaignUser {
			Campaign: "superbowl",
			LastName: "John",
			FirstName: "Doe",
			DeviceToken: uuid.New().String(),
			NativeLanguage: "EN",
			Membership: model.FreeTier,
			SortValue: "2",
		}
	}

	_, err := userRepo.InsertUsers(users)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("seeder run successfully")
}