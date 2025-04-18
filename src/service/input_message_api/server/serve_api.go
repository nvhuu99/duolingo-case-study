package main

import (
	ed "duolingo/event/event_data"
	eh "duolingo/event/event_handler"
	ep "duolingo/lib/event"
	mq "duolingo/lib/message_queue"
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service_container"
	model "duolingo/model"
	lc "duolingo/model/log/context"
	"duolingo/service/input_message_api/bootstrap"

	"log"

	"github.com/google/uuid"
)

var (
	container *sv.ServiceContainer
	publisher mq.Publisher
	server    *rest.Server
	event     *ep.EventPublisher
)

func input(request *rest.Request, response *rest.Response) {
	inputEvent := &ed.InputMessageRequest{
		OptId: uuid.NewString(),
		PushNoti: &model.PushNotiMessage{
			RelayFlag: model.ShouldRelay,
			Trace: &lc.TraceSpan{
				TraceId: uuid.NewString(),
			},
		},
		Request:  request,
		Response: response,
	}
	event.Notify(nil, eh.INP_MSG_REQUEST_BEGIN, inputEvent)
	defer event.Notify(nil, eh.INP_MSG_REQUEST_END, inputEvent)

	campaign := request.Path("campaign").Str()
	title := request.Input("title").Str()
	content := request.Input("content").Str()
	if title == "" {
		response.InvalidRequest("", map[string]string{
			"title": "title must not be empty",
		})
		return
	}
	if content == "" {
		response.InvalidRequest("", map[string]string{
			"content": "content must not be empty",
		})
		return
	}

	message := &model.InputMessage{
		MessageId: uuid.NewString(),
		Campaign:  campaign,
		Title:     title,
		Content:   content,
	}
	inputEvent.PushNoti.InputMessage = message

	err := publisher.Publish(message.Serialize())
	if err != nil {
		response.ServerErr("Failed to publish to message queue")
		return
	}
	response.Created("", message)
}

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	publisher = container.Resolve("mq.publisher").(mq.Publisher)
	server = container.Resolve("rest.server").(*rest.Server)
	event = container.Resolve("event.publisher").(*ep.EventPublisher)

	server.Router().Post("/campaign/{campaign}/message", input)

	log.Println("serving message input api")

	server.Serve()
}
