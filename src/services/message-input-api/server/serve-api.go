package main

import (
	"duolingo/common"
	"log"

	mq "duolingo/lib/message-queue"
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service-container"
	model "duolingo/model"
	"duolingo/services/message-input-api/bootstrap"
)

var (
	container 	*sv.ServiceContainer
	publisher 	mq.Publisher
	server		*rest.Server	
)

func input(request *rest.Request, response *rest.Response) {
	campaign := request.Path("campaign").Str()
	content := request.Input("content").Str()
	if content == "" {
		response.InvalidRequest("", map[string]string{
			"content": "content must not be empty",
		})
		return
	}

	message := model.NewInputMessage(campaign, content, false)
	err := publisher.Publish(message.Serialize())
	if err != nil {
		response.ServerErr("Failed to publish to message queue")
		return
	}

	response.Created("", message)
}

func main() {
	bootstrap.Run()

	container	= common.Container()
	publisher	= container.Resolve("mq.publisher").(mq.Publisher)
	server		= container.Resolve("rest.server").(*rest.Server)

	server.Router().Post("/campaign/{campaign}/message", input)

	log.Println("serving message input api")

	server.Serve()
}
