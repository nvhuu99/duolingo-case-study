package main

import (
	"context"
	config "duolingo/lib/config_reader/driver/reader/json"
	usr_repo "duolingo/repository/campaign_user"
	"flag"
	"log"
	"math"
	"path/filepath"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

func main() {
	// Read config
	conf := config.NewJsonReader(filepath.Join(".", "config"))

	total := flag.Int("total", 100, "total users to insert")
	campaign := flag.String("campaign", "", "campaign name")
	flag.Parse()
	if *campaign == "" {
		panic("campaign name must not be empty")
	}

	// Setup DB connection
	userRepo := usr_repo.NewUserRepo(
		context.Background(),
		conf.Get("db.campaign.name", ""),
	)
	err := userRepo.SetConnection(
		conf.Get("db.campaign.host", ""),
		conf.Get("db.campaign.port", ""),
		conf.Get("db.campaign.user", ""),
		conf.Get("db.campaign.password", ""),
	)
	if err != nil {
		panic(err)
	}

	batchSize := 1000
	numBatch := int(math.Ceil(float64(*total) / float64(batchSize)))

	// Initialize faker
	gofakeit.Seed(0)

	for i := range numBatch {
		count := batchSize
		if (i+1)*batchSize > *total {
			count = *total - i*batchSize
		}

		users := make([]*usr_repo.CampaignUser, count)

		for j := range count {
			membership, index := randomMembership()

			users[j] = &usr_repo.CampaignUser{
				Campaign:       *campaign,
				LastName:       gofakeit.LastName(),
				FirstName:      gofakeit.FirstName(),
				DeviceToken:    uuid.New().String(),
				NativeLanguage: gofakeit.LanguageAbbreviation(),
				Membership:     membership,
				SortValue:      int8(index),
				VerifiedAt:     time.Now().Add(-time.Duration(gofakeit.Number(10, 30)) * 24 * time.Hour),
			}
		}

		_, err := userRepo.InsertUsers(users)
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Printf("successfully created %v users for campaign %v", *total, *campaign)
}

// randomMembership returns a membership value and its index
func randomMembership() (usr_repo.Membership, int) {
	memberships := []usr_repo.Membership{
		usr_repo.Premium,
		usr_repo.Subscriber,
		usr_repo.FreeTier,
	}
	index := gofakeit.Number(0, len(memberships)-1)
	return memberships[index], index
}
