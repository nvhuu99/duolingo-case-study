package main

import (
	mq "duolingo/lib/message_queue"
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service_container"
	model "duolingo/model"
	"duolingo/service/message_input_api/bootstrap"
	"log"
)

var (
	container *sv.ServiceContainer
	publisher mq.Publisher
	server    *rest.Server
)

func input(request *rest.Request, response *rest.Response) {
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

	message := model.NewInputMessage(request.Id(), campaign, title, content, false)
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

	server.Router().Post("/campaign/{campaign}/message", input)

	log.Println("serving message input api")

	server.Serve()
}
