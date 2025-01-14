package bootstrap

import (
	"duolingo/common"
	"duolingo/lib/config-reader"
	mqp "duolingo/lib/message-queue"
	"duolingo/lib/message-queue/driver/rabbitmq"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	container = common.Container()
	ctx, _ = common.ServiceContext()
)

func bind() {
	info := getMQInfo("input_messages")
	if info == nil {
		panic("failed to get topic info from the message queue service api")
	}
	container.BindSingleton("publisher", func() any {
		publisher := rabbitmq.NewPublisher(ctx)
		publisher.SetTopicInfo(info)
		return publisher
	})
}

func Run() {
	common.SetupService()
	bind()
}

func getMQInfo(name string) *mqp.TopicInfo {
	conf, _ := container.Resolve("config").(config.ConfigReader)
	addr := conf.Get("services.mq_service_api", "")
	url := fmt.Sprintf("%v/topic/%v", addr, name)
	resp, httpErr := http.Get(url)

	var response struct {
		Success bool          `json:"success"`
		Data    mqp.TopicInfo `json:"data"`
		Message string        `json:"message"`
		Errors  any           `json:"errors"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)
	resp.Body.Close()
	
	if httpErr != nil {
		log.Println(httpErr.Error())
		return nil
	}
	
	if resp.StatusCode != http.StatusOK {
		log.Println(response.Message, "\n", response.Errors)
		return nil
	}	

	return &response.Data
}
