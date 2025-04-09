package main

import (
	lg "duolingo/lib/log"
	mq "duolingo/lib/message_queue"
	noti "duolingo/lib/notification"
	sv "duolingo/lib/service_container"
	"duolingo/model"
	ld "duolingo/model/log_detail"
	"duolingo/service/push_noti_sender/bootstrap"
	"encoding/json"
	"log"
)

var (
	container *sv.ServiceContainer
	consumer  mq.Consumer
	sender    noti.Sender
	logger    *lg.Logger
)

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	consumer = container.Resolve("mq.consumer").(mq.Consumer)
	sender = container.Resolve("noti.sender").(noti.Sender)
	logger = container.Resolve("logger").(*lg.Logger)

	log.Println("Notification worker started")

	consumer.Consume(make(chan bool, 1), func(jsonMsg string) mq.ConsumerAction {
		var message model.PushNotiMessage
		json.Unmarshal([]byte(jsonMsg), &message)

		result := sender.SendAll(message.Title, message.Content, message.DeviceTokens)

		logDetail := ld.SendNotificationDetail(&message, result)

		if result.Success {
			logger.Info("push notification success").Detail(logDetail).Write()
		} else {
			logger.Error("push notification failure", result.Error).Detail(logDetail).Write()
		}

		return mq.ConsumerAccept
	})
}
