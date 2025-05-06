package main

import (
	ed "duolingo/event/event_data"
	eh "duolingo/event/event_handler/service_opt"
	ep "duolingo/lib/event"
	mq "duolingo/lib/message_queue"
	noti "duolingo/lib/notification"
	sv "duolingo/lib/service_container"
	"duolingo/model"
	"duolingo/service/push_noti_sender/bootstrap"
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

var (
	container *sv.ServiceContainer
	consumer  mq.Consumer
	sender    noti.Sender
	event     *ep.EventPublisher
)

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	consumer = container.Resolve("mq.consumer").(mq.Consumer)
	sender = container.Resolve("noti.sender").(noti.Sender)
	event = container.Resolve("event.publisher").(*ep.EventPublisher)

	log.Println("notification worker started")

	consumer.Consume(make(chan bool, 1), func(body []byte) mq.ConsumerAction {
		pushNoti := new(model.PushNotiMessage)
		json.Unmarshal(body, pushNoti)
		return send(pushNoti)
	})
}

func send(pushNoti *model.PushNotiMessage) mq.ConsumerAction {
	result := &noti.Result{Success: true}
	sendEvent := &ed.SendPushNotification{OptId: uuid.NewString(), PushNoti: pushNoti}
	event.Notify(eh.SEND_PUSH_NOTI_BEGIN, sendEvent)
	defer event.Notify(eh.SEND_PUSH_NOTI_END, sendEvent)
	defer func() {
		sendEvent.Result = result
	}()

	tokenLimit := sender.GetTokenLimit()
	for i := 0; i < len(pushNoti.DeviceTokens); i += tokenLimit {
		end := min(i+tokenLimit, len(pushNoti.DeviceTokens))
		tokens := pushNoti.DeviceTokens[i:end]
		batchResult := sender.SendAll(
			pushNoti.InputMessage.Title,
			pushNoti.InputMessage.Content,
			tokens,
		)
		if !batchResult.Success {
			result.Success = false
			result.Error = batchResult.Error
			return mq.ConsumerRequeue
		} else {
			result.SuccessCount += batchResult.SuccessCount
			result.FailureCount += batchResult.FailureCount
		}
	}
	return mq.ConsumerAccept
}
