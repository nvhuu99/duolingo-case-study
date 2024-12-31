package main

import (
	"duolingo/lib/config-reader"
	mq "duolingo/lib/message-queue"
	container "duolingo/lib/service-container"
	"duolingo/services/duolingo-campaign-api/bootstrap"
	"fmt"
	"io"
	"log"
	"net/http"
)

func publishMessage(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		writer.WriteHeader(400)
		writer.Write([]byte("Invalid method"))
		return
	}

	campaign := request.PathValue("campaign")

	publisher := container.Resolve("publisher." + campaign).(mq.MessagePublisher)

	body, _ := io.ReadAll(request.Body)
	defer request.Body.Close()

	err := publisher.Publish(string(body))
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

	conf, _ := container.Resolve("config").(config.ConfigReader)

	addr := conf.Get("self.addr", "")

	http.HandleFunc("/campaign/{campaign}/message", publishMessage)

	log.Println("Serving Campaign Message Api: " + addr)

	http.ListenAndServe(addr, nil)
}
