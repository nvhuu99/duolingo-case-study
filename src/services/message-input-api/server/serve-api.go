package main

import (
	"duolingo/common"
	"duolingo/lib/config-reader"
	mq "duolingo/lib/message-queue"
	rest "duolingo/lib/rest-http"
	sv "duolingo/lib/service-container"
	model "duolingo/model"
	"duolingo/services/message-input-api/bootstrap"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var (
	container *sv.ServiceContainer
	conf      config.ConfigReader
)

func input(request *rest.Request, response *rest.Response) {
	campaign := request.Path("campaign").Str()
	content := request.Input("message").Str()
	if content == "" {
		response.InvalidRequest("", map[string]string{
			"message": "message must not be empty",
		})
		return
	}

	publisher := container.Resolve("mq.publisher").(mq.Publisher)

	message := model.InputMessage{
		Id:        uuid.New().String(),
		Content:   content,
		IsRelayed: false,
		Campaign:  campaign,
	}
	jsonMsg, _ := json.Marshal(message)

	err := publisher.Publish(string(jsonMsg))
	if err != nil {
		log.Println(err.Error())
		response.ServerErr("Failed to publish to message queue")
		return
	}

	response.Created(message)
}

func main() {
	bootstrap.Run()

	container = common.Container()
	conf = container.Resolve("config").(config.ConfigReader)
	addr := conf.Get("self.addr", "")
	if addr == "" {
		log.Fatal("message input api address is not provided")
	}

	go panicOnMessageQueueFailure()

	router := rest.NewRouter()
	router.Post("/campaign/{campaign}/message", input)

	log.Println("serving message input api at: " + addr)

	http.HandleFunc("/", router.Func())
	http.ListenAndServe(addr, nil)
}

func panicOnMessageQueueFailure() {
	errChan := container.Resolve("mq.err_chan").(chan error)
	err := <-errChan
	if err != nil {
		panic(err)
	}
}
