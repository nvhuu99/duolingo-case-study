package main

import (
	"duolingo/common"
	sv "duolingo/lib/service-container"
	"duolingo/lib/config-reader"
	mq "duolingo/lib/message-queue"
	rest "duolingo/lib/rest-http"
	model "duolingo/model"
	"duolingo/services/message-input-api/bootstrap"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var (
	container *sv.ServiceContainer
	conf config.ConfigReader
)

func input(request *rest.Request, response *rest.Response) {
	campaign := request.Path("campaign").Str()
	if campaign == "" {
		response.InvalidRequest("", map[string]string {
			"campaign": "campaign must not be empty",
		})
		return
	}
	content := request.Input("message").Str()
	if content == "" {
		response.InvalidRequest("", map[string]string {
			"message": "message must not be empty",
		})
		return
	}

	publisher := container.Resolve("publisher").(mq.MessagePublisher)

	message := model.InputMessage{
		Id: uuid.New().String(),
		Content: content,
		IsRelayed: false,
		Campaign: campaign,
	}
	jsonMsg, _ := json.Marshal(message)

	publisher.Connect() 
	defer publisher.Disconnect()
	
	err := publisher.Publish(string(jsonMsg))
	if err != nil {
		log.Println(err)
		response.ServerErr("Failed to publish to message queue")
		return
	}

	response.Created(message)
}

func main() {
	bootstrap.Run()

	container = common.Container()
	conf, _   = container.Resolve("config").(config.ConfigReader)

	router := rest.NewRouter()
	router.Post("/campaign/{campaign}/message", input)
	http.HandleFunc("/", router.Func())
	http.ListenAndServe(conf.Get("self.addr", ""), nil)
}
