package main

import (
	"context"
	"flag"
	"log"
	"time"

	"duolingo/dependencies"
	container "duolingo/libraries/dependencies_container"
	"duolingo/models"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/test/fixtures"
	"duolingo/test/fixtures/data"

	"github.com/google/uuid"
)

func main() {
	total := flag.Int("total", 10, "num of users")
	campaign := flag.String("campaign", data.TestCampaignPrimary, "campaign name")
	removeOldData := flag.Bool("remove", true, "remove old data before insert")
	flag.Parse()

	fixtures.SetTestConfigDir()
	dependencies.RegisterDependencies(context.Background())
	dependencies.BootstrapDependencies("test", []string{
		"common",
		"connections",
		"user_repo",
		"user_service",
	})
	factory := container.MustResolve[user_repo.UserRepoFactory]()
	repo := container.MustResolve[user_repo.UserRepository]()

	if *removeOldData {
		cmd := factory.MakeDeleteUsersCommand()
		cmd.SetFilterCampaign(*campaign)
		cmd.Build()
		err := repo.DeleteUsers(cmd)
		if err != nil {
			log.Println("failed to delete old data")
			return
		} else {
			log.Println("old data deleted")
		}
	}

	usrs := make([]*models.User, *total)
	for i := range *total {
		usrs[i] = &models.User{
			Id:        uuid.NewString(),
			Campaigns: []string{*campaign},
			Devices: []*models.UserDevice{
				{Token: uuid.NewString(), Platform: "android"},
				{Token: uuid.NewString(), Platform: "ios"},
			},
			EmailVerifiedAt: time.Now().UTC().Add(-1 * time.Hour),
		}
	}

	_, err := repo.InsertManyUsers(usrs)

	if err != nil {
		log.Println("insert test users failed - error:", err)
	} else {
		log.Printf("insert %v test users successfully\n", *total)
	}
}
