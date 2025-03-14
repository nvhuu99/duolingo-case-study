package main

import (
	"context"
	"duolingo/common"
	config "duolingo/lib/config-reader"
	mq "duolingo/lib/message-queue"
	sv "duolingo/lib/service-container"
	wd "duolingo/lib/work-distributor"
	"duolingo/model"
	db "duolingo/repository/campaign-db"
	"duolingo/services/noti-builder/bootstrap"
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

var (
	ctx         context.Context
	cancel      context.CancelFunc
	container   *sv.ServiceContainer
	conf		config.ConfigReader

	repo        *db.UserRepo
	distributor	wd.Distributor
	
	consumer			mq.Consumer
	inputMssgPublisher	mq.Publisher
	pushNotiPublisher	mq.Publisher
)

func main() {
	bootstrap.Run()

	container			= common.Container()
	ctx, cancel 		= common.ServiceContext()
	conf				= container.Resolve("config").(config.ConfigReader)
	repo				= container.Resolve("repo.campaign_user").(*db.UserRepo)
	distributor			= container.Resolve("distributor").(wd.Distributor)
	consumer			= container.Resolve("mq.consumer.input_messages").(mq.Consumer)
	inputMssgPublisher	= container.Resolve("mq.publisher.input_messages").(mq.Publisher)
	pushNotiPublisher	= container.Resolve("mq.publisher.push_noti_messages").(mq.Publisher)

	go cancelOnServicesFatalFailures()

	log.Println("Builder started")

	consumer.Consume(make(chan bool, 1), func(jsonMsg string) mq.ConsumerAction {
		var message model.InputMessage
		json.Unmarshal([]byte(jsonMsg), &message)
		// relay message
		if !message.IsRelayed {
			return relay(message)
		}
		// build noti messages		
		for {
			finished, action := build(message)
			if finished {
				return action
			}
		}
	})
}

func relay(message model.InputMessage) mq.ConsumerAction {
	// Register a new workload
	count, err := repo.CountUsers("superbowl")
	if err != nil {
		return mq.ConsumerRequeue
	}
	err = distributor.RegisterWorkLoad(&wd.Workload{
		Name: message.Id,
		NumOfUnits: count,
	})
	if err != nil {
		return mq.ConsumerRequeue
	}
	err = distributor.SwitchToWorkload(message.Id)
	if err != nil {
		return mq.ConsumerRequeue
	}
	// Build relay the message
	numOfBuilders := conf.GetInt("self.num_of_builders", 1)
	batch := make([]string, numOfBuilders)
	for i := 0; i < numOfBuilders; i++ {
		serialized, _ := json.Marshal(model.InputMessage {
			Id: message.Id,
			Content: message.Content,
			Campaign: message.Campaign,
			IsRelayed: true,
		})
		batch[i] = string(serialized)
	}
	// Start relaying
	i := 0
	for {
		select {
		case <-ctx.Done():
			return mq.ConsumerRequeue
		default:
		}
		if i == len(batch) {
			return mq.ConsumerAccept
		}
		if err := inputMssgPublisher.Publish(batch[i]); err != nil {
			return mq.ConsumerRequeue
		}
		i++
	}
}

func build(message model.InputMessage) (bool, mq.ConsumerAction) {
	var err error
	var assignment *wd.Assignment
	var users []*model.CampaignUser

	done := func() (bool, mq.ConsumerAction) {
		if assignment != nil {
			distributor.Commit(assignment.Id)
			return false, mq.ConsumerAccept
		}
		return true, mq.ConsumerAccept
	}

	abort := func() (bool, mq.ConsumerAction) {
		if assignment != nil {
			distributor.RollBack(assignment.Id)
		}
		return false, mq.ConsumerRequeue
	}

	for {
		select {
		case <-ctx.Done():
			return abort()
		default:
		}

		assignment, err = distributor.Next()
		if err != nil {
			return abort()
		}
		if assignment == nil {
			return done()
		}

		users, err = repo.UsersList(&db.ListUserOptions{
			Campaign:   message.Campaign,
			Skip:       assignment.Start - 1,
			Limit:      assignment.End - assignment.Start + 1,
			CursorMode: false,
		})
		if err != nil {
			return abort()
		}

		i := 0
		for {
			if i == len(users) {
				return done()
			}
			data := &model.PushNotiMessage{
				Id:          uuid.New().String(),
				Content:     message.Content,
				DeviceToken: users[i].DeviceToken,
			}
			noti, _ := json.Marshal(data)
			if err = pushNotiPublisher.Publish(string(noti)); err != nil {
				return abort()
			}
			distributor.Progress(message.Id, assignment.Progress+1)
			assignment.Progress++
			i++
		}
	}
}

func cancelOnServicesFatalFailures() {
	errChan := container.Resolve("err_chan").(chan error)
	err := <-errChan
	if err != nil {
		cancel()
	}
}
