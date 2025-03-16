package main

import (
	"context"
	"duolingo/common"
	mq "duolingo/lib/message-queue"
	sv "duolingo/lib/service-container"
	noti "duolingo/lib/notification"
	"duolingo/model"
	"duolingo/services/push-noti-sender/bootstrap"
	"encoding/json"
	"log"
)

var (
	cancel      context.CancelFunc
	container   *sv.ServiceContainer
	consumer	mq.Consumer
	sender		noti.Sender
)

func main() {
	bootstrap.Run()

	container			= common.Container()
	_, cancel			= common.ServiceContext()
	consumer			= container.Resolve("mq.consumer").(mq.Consumer)
	sender				= container.Resolve("noti.sender").(noti.Sender)

	go cancelOnServicesFatalFailures()

	log.Println("Notification worker started")

	consumer.Consume(make(chan bool, 1), func(jsonMsg string) mq.ConsumerAction {
		var message model.PushNotiMessage
		log.Println("sender:receive:", message)
		
		json.Unmarshal([]byte(jsonMsg), &message)
		
		result, err := sender.SendAll(&noti.Message{
			Title: "",
			Body: message.Content,
		}, message.DeviceTokens)

		if err != nil {
			log.Println("noti:err:", err)
		} else {
			log.Printf("noti:ok:count:%v:failure_count:%v", len(message.DeviceTokens), result.FailureCount)
		}

		return mq.ConsumerAccept
	})
}

func cancelOnServicesFatalFailures() {
	errChan := container.Resolve("err_chan").(chan error)
	err := <-errChan
	if err != nil {
		cancel()
	}
}
