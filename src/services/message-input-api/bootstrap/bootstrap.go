package bootstrap

import (
	"context"
	"duolingo/common"
	"duolingo/lib/config-reader"
	mq "duolingo/lib/message-queue"
	"duolingo/lib/message-queue/driver/rabbitmq"
	sv "duolingo/lib/service-container"
	"time"
)

var (
	container *sv.ServiceContainer
	ctx       context.Context
	infra     config.ConfigReader

	graceTimeOut   = 100 * time.Millisecond
	connTimeOut    = 10 * time.Second
	heartBeat      = 10 * time.Second
	declareTimeOut = 10 * time.Second
	writeTimeOut   = 5 * time.Second
)

func bind() {
	container.BindSingleton("mq.err_chan", func() any {
		return make(chan error, 2)
	})

	container.BindSingleton("mq.manager", func() any {
		manager := rabbitmq.NewRabbitMQManager(ctx)
		manager.
			WithOptions(nil).
			WithGraceTimeOut(graceTimeOut).
			WithConnectionTimeOut(connTimeOut).
			WithHearBeat(heartBeat).
			WithKeepAlive(true)
		manager.
			UseConnection(
				infra.Get("mq.host", ""),
				infra.Get("mq.port", ""),
				infra.Get("mq.user", ""),
				infra.Get("mq.pwd", ""),
			)
		manager.
			Connect()

		return manager
	})

	container.BindSingleton("mq.topology", func() any {
		manager, _ := container.Resolve("mq.manager").(mq.Manager)
		errChan, _ := container.Resolve("mq.err_chan").(chan error)

		topology := rabbitmq.
			NewRabbitMQTopology("campaign_messages_topology", ctx)
		topology.
			UseManager(manager)
		topology.
			NotifyError(errChan)
		topology.
			WithOptions(nil).
			WithGraceTimeOut(graceTimeOut).
			WithDeclareTimeOut(declareTimeOut).
			WithQueuesPurged(false)
		topology.
			Topic("campaign_messages").Queue("input_messages").Bind("input_messages")

		return topology
	})

	container.BindSingleton("mq.publisher", func() any {
		manager, _ := container.Resolve("mq.manager").(mq.Manager)
		errChan, _ := container.Resolve("mq.err_chan").(chan error)

		publisher := rabbitmq.
			NewPublisher("input_messages_publisher", ctx)
		publisher.
			UseManager(manager)
		publisher.
			NotifyError(errChan)
		publisher.
			WithOptions(nil).
			WithGraceTimeOut(graceTimeOut).
			WithWriteTimeOut(writeTimeOut).
			WithTopic("campaign_messages").
			WithDirectDispatch("input_messages")

		return publisher
	})
}

func boot() {
}

func Run() {
	common.SetupService()

	container = common.Container()
	infra = container.Resolve("config.infra").(config.ConfigReader)
	ctx, _ = common.ServiceContext()

	bind()
	boot()
}
