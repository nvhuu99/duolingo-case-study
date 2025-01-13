package main

import (
	"duolingo/common"
	"duolingo/lib/config-reader"
	mq "duolingo/lib/message-queue"
	rest "duolingo/lib/rest-http"
	model "duolingo/model"
	"duolingo/services/message-input-api/bootstrap"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	container = common.Container()
	conf, _   = container.Resolve("config").(config.ConfigReader)
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
		Id: "message." + strconv.FormatInt(time.Now().UnixMicro(), 10),
		Content: content,
		IsRelayed: false,
		Campaign: campaign,
	}
	jsonMsg, _ := json.Marshal(message)

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

	router := rest.Router{}
	router.Post("/campaign/{campaign}/message", input)
	http.HandleFunc("/", router.Func())
	http.ListenAndServe(conf.Get("self.addr", ""), nil)
}
