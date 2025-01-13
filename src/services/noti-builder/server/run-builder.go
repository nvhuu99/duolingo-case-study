package main

import (
	"duolingo/common"
	bm "duolingo/lib/batch-manager"
	"duolingo/lib/config-reader"
	mqp "duolingo/lib/message-queue"
	"duolingo/model"
	"duolingo/services/noti-builder/bootstrap"
	"encoding/json"
	"log"
	"strconv"
	"time"
	db "duolingo/repository/campaign-db"
)

var (
	ctx, cancel = common.ServiceContext()
	container = common.Container()
	queue = container.Resolve("queue_info").(*mqp.QueueInfo)
	consumer = container.Resolve("consumer.input_messages").(mqp.MessageConsumer)
	ipPublisher = container.Resolve("publisher.input_messages").(mqp.MessagePublisher)
	pnPublisher = container.Resolve("publisher.push_noti_messages").(mqp.MessagePublisher)
	manager = container.Resolve("manager.noti_builder").(bm.BatchManager)
	repo = container.Resolve("repo.campaign_user").(*db.UserRepo)
	conf = container.Resolve("config").(config.ConfigReader)
)

func main() {
	bootstrap.Run()

	consumer.Consume(func (jsonMsg string) bool {
		var message model.InputMessage 
		json.Unmarshal([]byte(jsonMsg), &message)

		if !message.IsRelayed {
			if !relay(message) {
				cancel()
				return false
			}
		}

		if !build(message) {
			cancel()
			return false
		}

		return true
	})

	<-ctx.Done()
}

func relay(message model.InputMessage) bool {
	if err := manager.NewBatch(message.Id); err != nil {
		log.Println(err)
		return false
	}

	batch := make([]string, queue.TotalConsumer)
	for i := 0; i <= queue.TotalConsumer; i++ {
		clone := message
		clone.IsRelayed = true
		serialized, _ := json.Marshal(clone)
		batch[i] = string(serialized)
	}

	for _, mssg := range batch {
		if err := ipPublisher.Publish(mssg); err != nil {
			log.Println(err.Error())
			return false
		}
		return true
	}

	return true
}

func build(message model.InputMessage) bool {
	batch, err := manager.Next(message.Id)
	if err != nil {
		return false
	}

	go func() {
		time.Sleep(5 * time.Second)
		manager.RollBack(message.Id, batch.Id)
	}()

	var notiErr error
	progress := batch.Start
	
	handler := func (user *model.CampaignUser) bool {
		data := &model.PushNotiMessage{
			Id: "noti." + strconv.FormatInt(time.Now().UnixMicro(), 10),
			Content: message.Content,
			DeviceToken: user.DeviceToken,
		}
		noti, _ := json.Marshal(data)

		if err := pnPublisher.Publish(string(noti)); err != nil {
			notiErr = err
			return false
		}

		manager.Progress(message.Id, batch.Id, progress)
		progress++

		return true
	}

	if batch.HasFailed {
		batch.Start = batch.Progress
	}
	repo.CampaignUsersList(&db.ListUserOptions{
		Campaign: message.Campaign,
		Skip: batch.Start - 1,
		Limit: batch.End - batch.Start + 1,
		CursorMode: true,
		CursorFunc: handler,
	})

	if notiErr != nil {
		log.Println(notiErr)
		manager.RollBack(message.Id, batch.Id)
		return false
	}

	manager.Commit(message.Id, batch.Id)

	return true
}
