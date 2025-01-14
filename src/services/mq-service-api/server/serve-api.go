package main

import (
	"net/http"

	"duolingo/common"
	"duolingo/lib/config-reader"
	sv "duolingo/lib/service-container"
	mqp "duolingo/lib/message-queue"
	"duolingo/services/mq-service-api/bootstrap"
	rest "duolingo/lib/rest-http"
)

var (
	container *sv.ServiceContainer
	conf config.ConfigReader
)

func topicInfo(request *rest.Request, response *rest.Response) {
	name := request.Path("name").Str()	
	mq, _ := container.Resolve("topic." + name).(mqp.MessageQueueService)
	if mq == nil {
		response.NotFound("topic does not exist")
		return
	}

	info := mq.GetTopicInfo()

	response.Ok(info)
}

func queueInfo(request *rest.Request, response *rest.Response) {
	name := request.Path("name").Str()	
	queue := request.Path("queue").Str()
	mq, _ := container.Resolve("topic." + name).(mqp.MessageQueueService)
	if mq == nil {
		response.NotFound("topic does not exist")
		return
	}

	info, _ := mq.GetQueueInfo(name + "." + queue)
	if info == nil {
		response.NotFound("queue does not exist")
		return
	}

	response.Ok(info)
}

func registerConsumer(request *rest.Request, response *rest.Response) {
	name := request.Path("name").Str()
	worker := request.Path("consumer").Str()

	mq, _ := container.Resolve("topic." + name).(mqp.MessageQueueService)
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

	container = common.Container()
	conf, _ = container.Resolve("config").(config.ConfigReader)
	addr := conf.Get("self.addr", "")

	router := rest.NewRouter()
	router.Get("/topic/{name}", topicInfo)
	router.Get("/topic/{name}/queue/{queue}", queueInfo)
	router.Post("/topic/{name}/consumer/{consumer}", registerConsumer)
	http.HandleFunc("/", router.Func())
	http.ListenAndServe(addr, nil)
}
