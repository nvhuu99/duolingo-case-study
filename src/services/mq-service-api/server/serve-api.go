package main

import (
	"net/http"

	"duolingo/common"
	"duolingo/lib/config-reader"
	mqp "duolingo/lib/message-queue"
	"duolingo/services/mq-service-api/bootstrap"
	rest "duolingo/lib/rest-http"
)

var (
	container = common.Container()
	conf, _ = container.Resolve("config").(config.ConfigReader)
)

func topicInfo(request *rest.Request, response *rest.Response) {
	name := request.Path("name").Str()	
	mq := container.Resolve("topic." + name).(mqp.MessageQueueService)
	if mq == nil {
		response.NotFound("topic does not exist")
		return
	}

	info := mq.GetTopicInfo()

	response.Ok(info)
}

func registerConsumer(request *rest.Request, response *rest.Response) {
	name := request.Path("name").Str()
	worker := request.Path("consumer").Str()

	mq := container.Resolve("topic." + name).(mqp.MessageQueueService)
	if mq == nil {
		response.NotFound("topic does not exist")
		return
	}

	queueInfo, err := mq.RegisterConsumer(worker)
	if err != nil {
		response.ServerErr(err.Error())
		return
	}

	response.Ok(queueInfo)
}

func main() {
	bootstrap.Run()

	addr := conf.Get("self.addr", "")
	router := rest.Router{}
	router.Get("/topic/{name}/info", topicInfo)
	router.Post("/topic/{name}/consumer/{consumer}", registerConsumer)
	http.HandleFunc("/", router.Func())
	http.ListenAndServe(addr, nil)
}
