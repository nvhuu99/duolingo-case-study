package bootstrap

import (
	"duolingo/common"
	"duolingo/lib/config-reader"
	mqp "duolingo/lib/message-queue"
	"duolingo/lib/message-queue/driver/rabbitmq"
	"duolingo/lib/service-container"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func bind() {
	ctx, _ := common.ServiceContext()
	container.BindSingleton("publisher.superbowl", func() any {
		info := getMQInfo("superbowl")
		publisher := mqp.MessagePublisher(rabbitmq.NewPublisher(ctx))
		publisher.SetTopicInfo(info)
		return publisher
	})
}

func Run() {
	common.SetupService()
	bind()
}

func getMQInfo(campaign string) mqp.TopicInfo {
	conf, _ := container.Resolve("config").(config.ConfigReader)
	url := fmt.Sprintf("%v/topic/%v/info", conf.Get("services.mqServiceApi.addr", ""), campaign)

	resp, err := http.Get(url)
	if err != nil {
		panic("failed to get campaign topic info\n" + err.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		panic("failed to get campaign topic info\n" + string(body))
	}

	var response struct {
		Success bool          `json:"success"`
		Data    mqp.TopicInfo `json:"data"`
	}
	json.Unmarshal(body, &response)

	return response.Data
}
