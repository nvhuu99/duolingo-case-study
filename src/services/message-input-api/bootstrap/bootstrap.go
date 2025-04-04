package bootstrap

import (
	"context"
	"duolingo/common"
	"duolingo/lib/config_reader"
	rest "duolingo/lib/rest_http"
	md "duolingo/services/message-input-api/server/middleware"
	mq "duolingo/lib/message-queue"
	"duolingo/lib/message-queue/driver/rabbitmq"
	sv "duolingo/lib/service-container"
	log "duolingo/lib/log"

	"time"
)

var (
	container	*sv.ServiceContainer
	ctx			context.Context
	infra		config.ConfigReader
	conf		config.ConfigReader

	graceTimeOut   = 100 * time.Millisecond
	connTimeOut    = 10 * time.Second
	heartBeat      = 10 * time.Second
	declareTimeOut = 10 * time.Second
	writeTimeOut   = 5 * time.Second
)

func Run() {
	common.SetupService()

	container	= common.Container()
	infra		= container.Resolve("config.infra").(config.ConfigReader)
	conf		= container.Resolve("config").(config.ConfigReader)
	ctx, _		= common.ServiceContext()

	bindLogger()
	bindRestHttp()
	bindMessageQueue()
}

func bindLogger() {
	container.BindSingleton("log.server", func() any {
		rotation := time.Duration(conf.GetInt("self.log.rotation", 86400))
		bufferSize := conf.GetInt("self.log.buffer.size", 2)
		bufferCount := conf.GetInt("self.log.buffer.max_count", 1000)
		return log.NewLogger(ctx, "server").
			UseJsonFormat().
			AddLocalWriter(common.Dir("storage/log"), bufferSize, bufferCount, rotation)
	})
}

func bindRestHttp() {
	container.BindSingleton("rest.server", func() any {
		logger := container.Resolve("log.server").(*log.Logger)
		server := rest.
			NewServer(conf.Get("self.server.addr", ":80")).
			WithMiddlewares("response", &md.LogHandledRequest{ Logger: logger })

		return server
	})
}

func bindMessageQueue() {
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