package main

import (
	ed "duolingo/event/event_data"
	eh "duolingo/event/event_handler"
	ep "duolingo/lib/event"
	lg "duolingo/lib/log"
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
	logger    *lg.Logger
	event     *ep.EventPublisher
)

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	consumer = container.Resolve("mq.consumer").(mq.Consumer)
	sender = container.Resolve("noti.sender").(noti.Sender)
	logger = container.Resolve("logger").(*lg.Logger)
	event = container.Resolve("event.publisher").(*ep.EventPublisher)

	log.Println("notification worker started")

	consumer.Consume(make(chan bool, 1), func(body []byte) mq.ConsumerAction {
		pushNoti := new(model.PushNotiMessage)
		json.Unmarshal(body, pushNoti)
		return send(pushNoti)
	})
}

func send(pushNoti *model.PushNotiMessage) mq.ConsumerAction {
	var result *noti.Result

	sendEvent := &ed.SendPushNotification{OptId: uuid.NewString(), PushNoti: pushNoti}
	event.Notify(nil, eh.SEND_PUSH_NOTI_BEGIN, sendEvent)
	defer event.Notify(nil, eh.SEND_PUSH_NOTI_END, sendEvent)
	defer func() {
		sendEvent.Result = result
	}()

	result = sender.SendAll(
		pushNoti.InputMessage.Title,
		pushNoti.InputMessage.Content,
		pushNoti.DeviceTokens,
	)

	if !result.Success {
		return mq.ConsumerRequeue
	}
	return mq.ConsumerAccept
}
