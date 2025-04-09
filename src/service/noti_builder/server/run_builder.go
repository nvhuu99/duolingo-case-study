package main

import (
	"context"
	config "duolingo/lib/config_reader"
	lg "duolingo/lib/log"
	mq "duolingo/lib/message_queue"
	sv "duolingo/lib/service_container"
	wd "duolingo/lib/work_distributor"
	"duolingo/model"
	ld "duolingo/model/log_detail"
	db "duolingo/repository/campaign_db"
	"duolingo/service/noti_builder/bootstrap"
	"encoding/json"
	"log"
)

var (
	ctx       context.Context
	container *sv.ServiceContainer
	conf      config.ConfigReader

	repo        *db.UserRepo
	distributor wd.Distributor
	logger      *lg.Logger

	consumer           mq.Consumer
	inputMssgPublisher mq.Publisher
	pushNotiPublisher  mq.Publisher
)

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	ctx = context.Background()
	conf = container.Resolve("config").(config.ConfigReader)
	repo = container.Resolve("repo.campaign_user").(*db.UserRepo)
	distributor = container.Resolve("distributor").(wd.Distributor)
	consumer = container.Resolve("mq.consumer.input_messages").(mq.Consumer)
	logger = container.Resolve("log.server").(*lg.Logger)
	inputMssgPublisher = container.Resolve("mq.publisher.input_messages").(mq.Publisher)
	pushNotiPublisher = container.Resolve("mq.publisher.push_noti_messages").(mq.Publisher)

	log.Println("builder started")

	consumer.Consume(make(chan bool, 1), func(jsonMsg string) mq.ConsumerAction {
		var message model.InputMessage
		json.Unmarshal([]byte(jsonMsg), &message)
		// relay message
		if !message.IsRelayed {
			return relay(message)
		}
		// build noti messages
		return build(message)
	})
}

func relay(message model.InputMessage) mq.ConsumerAction {
	failed := func(failErr error) mq.ConsumerAction {
		logger.Error("relay failure", failErr).Detail(ld.RelayInputMessageDetail(&message, 0)).Write()
		return mq.ConsumerRequeue
	}

	skipped := func(reason string) mq.ConsumerAction {
		logger.Info("message skipped: " + reason).Detail(ld.SkipInputMessageDetail(&message, reason)).Write()
		return mq.ConsumerAccept
	}

	completed := func(relayedTotal int) mq.ConsumerAction {
		logger.Info("relay success").Detail(ld.RelayInputMessageDetail(&message, relayedTotal)).Write()
		return mq.ConsumerAccept
	}

	// Register a new workload
	count, err := repo.CountUsers(message.Campaign)
	if err != nil {
		return failed(err)
	}
	if count == 0 {
		// skip message due to no campaign user
		return skipped("no campaign user found")
	}
	err = distributor.RegisterWorkLoad(&wd.Workload{
		Name:       message.Id,
		NumOfUnits: count,
	})
	if err != nil {
		return failed(err)
	}
	// Build relay the message
	numOfBuilders := conf.GetInt("noti_builder.server.num_of_builders", 1)
	batch := make([]string, numOfBuilders)
	for i := 0; i < numOfBuilders; i++ {
		serialized, _ := json.Marshal(model.InputMessage{
			Id:        message.Id,
			RequestId: message.RequestId,
			Title:     message.Title,
			Content:   message.Content,
			Campaign:  message.Campaign,
			IsRelayed: true,
		})
		batch[i] = string(serialized)
	}
	// Start relaying
	i := 0
	for {
		select {
		case <-ctx.Done():
			return failed(nil)
		default:
		}
		if i == len(batch) {
			return completed(numOfBuilders)
		}
		if err := inputMssgPublisher.Publish(batch[i]); err != nil {
			return failed(err)
		}
		i++
	}
}

func build(message model.InputMessage) mq.ConsumerAction {
	var err error
	var assignment *wd.Assignment
	var workload *wd.Workload
	var users []*model.CampaignUser

	abort := func(abortErr error) mq.ConsumerAction {
		if assignment != nil {
			distributor.RollBack(assignment.Id)
		}
		logger.Error("build failed, assignment rollbacked", abortErr).Detail(ld.BuildNotificationDetail(&message, workload, assignment)).Write()
		return mq.ConsumerRequeue
	}

	complete := func() mq.ConsumerAction {
		logger.Info("build completed, workload completed").Detail(ld.BuildNotificationDetail(&message, workload, assignment)).Write()
		return mq.ConsumerAccept
	}

	workload, err = distributor.SwitchToWorkload(message.Id)
	if err != nil {
		return abort(err)
	}

	for {
		select {
		case <-ctx.Done():
			return abort(nil)
		default:
		}

		assignment, err = distributor.Next()
		if err != nil {
			if assignment == nil {
				return complete()
			}
			return abort(err)
		}

		users, err = repo.UsersList(&db.ListUserOptions{
			Campaign:   message.Campaign,
			Skip:       assignment.Start - 1,
			Limit:      assignment.End - assignment.Start + 1,
			CursorMode: false,
		})
		if err != nil {
			return abort(err)
		}

		deviceTokens := make([]string, len(users))
		for i, user := range users {
			deviceTokens[i] = user.DeviceToken
		}

		pushNoti := model.NewPushNotiMessage(
			message.RequestId,
			message.Title,
			message.Content,
			deviceTokens,
		)
		if err = pushNotiPublisher.Publish(pushNoti.Serialize()); err != nil {
			return abort(err)
		}
		distributor.Progress(message.Id, assignment.End)
		distributor.Commit(assignment.Id)

		logger.Info("build success, assignment commited").Detail(ld.BuildNotificationDetail(&message, workload, assignment)).Write()
	}
}
