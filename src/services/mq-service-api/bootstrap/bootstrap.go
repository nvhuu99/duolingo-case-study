package bootstrap

import (
	"duolingo/common"
	"duolingo/lib/config-reader"
	mqp "duolingo/lib/message-queue"
	rabbitmq "duolingo/lib/message-queue/driver/rabbitmq"
	"duolingo/lib/service-container"
)

func bind() {
	ctx, _ := common.ServiceContext()
	// Bind Superbowl MQ Topic using RabbitMQ driver
	container.BindSingleton("topic.superbowl", func() any {
		return mqp.MessageQueueService(rabbitmq.NewMQService(ctx))
	})
}

func boot() {
	conf, _ := container.Resolve("config").(config.ConfigReader)
	infra, _ := container.Resolve("config.infra").(config.ConfigReader)
	
	// Setup Superbowl MQ Topic
	mq, _ := container.Resolve("topic.superbowl").(mqp.MessageQueueService)
	mq.UseConnection(
		infra.Get("mq.conn.host", ""),
		infra.Get("mq.conn.port", ""),
		infra.Get("mq.conn.user", ""),
		infra.Get("mq.conn.pwd", ""),
	)
	mq.SetTopic(conf.Get("mq.superbowl.topic", ""))
	mq.SetNumberOfQueue(conf.GetInt("mq.superbowl.numOfQueue", 1))
	mq.SetQueueConsumerLimit(conf.GetInt("mq.superbowl.queueConsumerLimit", 1))
	mq.SetDistributeMethod(mqp.DistributeMethod(conf.Get("mq.superbowl.method", "")))
	err := mq.Publish()
	if err != nil {
		mq.Shutdown()
		panic("failed to setup message queue topic 'campaign'")
	}
}

func Run() {
	common.SetupService()
	bind()
	boot()
}
