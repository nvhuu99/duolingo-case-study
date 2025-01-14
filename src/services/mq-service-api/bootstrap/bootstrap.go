package bootstrap

import (
	"context"
	"duolingo/common"
	"duolingo/lib/config-reader"
	mqp "duolingo/lib/message-queue"
	rabbitmq "duolingo/lib/message-queue/driver/rabbitmq"
	sv "duolingo/lib/service-container"
	"log"
)

var (
	ctx context.Context 
	container *sv.ServiceContainer
	conf config.ConfigReader
	infra config.ConfigReader
)

func bind() {
	topics := conf.GetArr("mq.topics", []string{})
	for _, tp := range topics {
		container.BindSingleton("topic." + tp, func() any {
			return rabbitmq.NewMQService(ctx)
		})
	}
}

func boot() {
	topics := conf.GetArr("mq.topics", []string{})
	for _, tp := range topics {
		mq, _ := container.Resolve("topic." + tp).(mqp.MessageQueueService)
		mq.UseConnection(
			infra.Get("mq.host", ""),
			infra.Get("mq.port", ""),
			infra.Get("mq.user", ""),
			infra.Get("mq.pwd", ""),
		)
		mq.SetTopic(conf.Get("mq." + tp + ".topic", ""))
		mq.SetNumberOfQueue(conf.GetInt("mq." + tp + ".num_of_queue", 1))
		mq.SetQueueConsumerLimit(conf.GetInt("mq." + tp + ".queue_consumer_limit", 1))
		mq.SetDistributeMethod(mqp.DistributeMethod(conf.Get("mq." + tp + ".method", "")))
		err := mq.Publish()
		if err != nil {
			mq.Shutdown()
			log.Fatalf("failed to setup message queue topic '%v'\n%v", tp, err)
		}
	}
}

func Run() {
	common.SetupService()
	ctx, _ = common.ServiceContext()
	container = common.Container()
	conf, _ = container.Resolve("config").(config.ConfigReader)
	infra, _ = container.Resolve("config.infra").(config.ConfigReader)
	
	bind()

	boot()
}
