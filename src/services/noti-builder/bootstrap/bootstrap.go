package bootstrap

import (
	"duolingo/common"
	redismanager "duolingo/lib/batch-manager/driver/redis"
	"duolingo/lib/config-reader"
	db "duolingo/repository/campaign-db"
	mqp "duolingo/lib/message-queue"
	"duolingo/lib/message-queue/driver/rabbitmq"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	container = common.Container()
	ctx, _ = common.ServiceContext()
	conf, _ = container.Resolve("config").(config.ConfigReader)
	infra, _ = container.Resolve("config.infra").(config.ConfigReader)
)

func bind() {
	container.Bind("queue_info", func() any {
		name := "consumer.input_messages." + strconv.FormatInt(time.Now().UnixMicro(), 10) 
		queueInfo := registerConsumer("input_messages", name)
		if queueInfo == nil {
			panic("failed to register input_messages consumer to the message queue service api")
		}
		return queueInfo
	})
	
	container.BindSingleton("consumer.input_messages", func() any {
		queueInfo := container.Resolve("queue_info").(*mqp.QueueInfo)
		consumer := rabbitmq.NewConsumer(ctx)
		consumer.SetQueueInfo(queueInfo)
		return consumer
	})

	container.BindSingleton("publisher.input_messages", func() any {
		info := getMQInfo("input_messages")
		if info == nil {
			panic("failed to get topic info from the message queue service api")
		}
		publisher := mqp.MessagePublisher(rabbitmq.NewPublisher(ctx))
		publisher.SetTopicInfo(info)
		return publisher
	})

	container.BindSingleton("publisher.push_noti_messages", func() any {
		info := getMQInfo("push_noti_messages")
		if info == nil {
			panic("failed to get topic info from the message queue service api")
		}
		publisher := mqp.MessagePublisher(rabbitmq.NewPublisher(ctx))
		publisher.SetTopicInfo(info)
		return publisher
	})

	container.BindSingleton("repo.campaign_user", func() any {
		repo := db.NewUserRepo(ctx, infra.Get("db.campaign.name", ""))
		repo.SetConnection(
			infra.Get("db.campaign.host", ""),
			infra.Get("db.campaign.port", ""),
			infra.Get("db.campaign.user", ""),
			infra.Get("db.campaign.password", ""),
		)
		return repo
	})

	container.BindSingleton("manager.noti_builder", func() any {
		repo := container.Resolve("repo.campaign_ser").(*db.UserRepo)
		count, err := repo.CountUsers("superbowl")
		if err != nil {
			log.Println(err.Error())
			panic("failed to setup batch manager")
		}

		size := infra.GetInt("batch.noti_builder.batch_size", 1000)
		manager := redismanager.GetBatchManager(ctx, "noti_builder", 1, count, size)
		manager.SetConnection(
			infra.Get("redis.host", ""),
			infra.Get("redis.port", ""),
		)
		// must not call reset as the batch manager service already did
		// manager.Reset() 

		return manager
	})
}

func Run() {
	common.SetupService()
	bind()
}

func registerConsumer(campaign string, consumer string) *mqp.QueueInfo {
	addr := conf.Get("services.mq_service_api", "")
	url := fmt.Sprintf("%v/topic/%v/consumer/%v", addr, campaign, consumer)
	resp, httpErr := http.Post(url, "", nil)

	var response struct {
		Success bool          `json:"success"`
		Data    mqp.QueueInfo `json:"data"`
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

func getMQInfo(name string) *mqp.TopicInfo {
	addr := conf.Get("services.mq_service_api", "")
	url := fmt.Sprintf("%v/topic/%v/info", addr, name)
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
