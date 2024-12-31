package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"duolingo/lib/config-reader"
	container "duolingo/lib/service-container"
	"duolingo/services/mq-service-api/bootstrap"
	mqp "duolingo/lib/message-queue"
)

func getTopicInfo(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		writer.WriteHeader(400)
		writer.Write([]byte("Invalid method"))
		return
	}

	name := request.PathValue("name")
	mq := container.Resolve("topic." + name).(mqp.MessageQueueService)
	if mq == nil {
		writer.WriteHeader(404)
		writer.Write([]byte("Topic not found"))
		return
	}

	info := mq.GetTopicInfo()
	data := map[string]any{
		"success": true,
		"data":    info,
	}
	json, _ := json.Marshal(data)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	writer.Write(json)
}

func getQueueInfo(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		writer.WriteHeader(400)
		writer.Write([]byte("Invalid method"))
		return
	}

	name := request.PathValue("name")
	queue := request.PathValue("queue")
	mq := container.Resolve("topic." + name).(mqp.MessageQueueService)
	if mq == nil {
		writer.WriteHeader(404)
		writer.Write([]byte("Topic not found"))
		return
	}

	info, err := mq.GetQueueInfo(queue)
	if err != nil {
		jsonErr := fmt.Sprintf(`{ "success": "false", "errors": ["%v"] }`, err.Error())
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(500)
		writer.Write([]byte(jsonErr))
		return
	}

	data := map[string]any{
		"success": true,
		"data":    info,
	}
	json, _ := json.Marshal(data)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	writer.Write(json)
}

func registerConsumer(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		writer.WriteHeader(400)
		writer.Write([]byte("Invalid method"))
		return
	}

	name := request.PathValue("name")
	worker := request.PathValue("consumer")

	mq := container.Resolve("topic." + name).(mqp.MessageQueueService)
	if mq == nil {
		writer.WriteHeader(404)
		writer.Write([]byte("Topic not found"))
		return
	}

	err := mq.RegisterConsumer(worker)
	if err != nil {
		jsonErr := fmt.Sprintf(`{ "success": "false", "errors": ["%v"] }`, err.Error())
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(500)
		writer.Write([]byte(jsonErr))

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	writer.Write([]byte(`{ "success": "true", "data": {}`))
}

func main() {
	bootstrap.Run()

	// GET - get topic info
	http.HandleFunc("/topic/{name}/info", getTopicInfo)
	// GET - get queue info
	http.HandleFunc("/topic/{name}/queue/{queue}/info", getQueueInfo)
	// POST - register topic queue consumer
	http.HandleFunc("/topic/{name}/consumer/{consumer}", registerConsumer)

	conf, _ := container.Resolve("config").(config.ConfigReader)
	addr := conf.Get("self.addr", "")

	log.Println("Serving MQ Service API: " + addr)

	http.ListenAndServe(addr, nil)
}
