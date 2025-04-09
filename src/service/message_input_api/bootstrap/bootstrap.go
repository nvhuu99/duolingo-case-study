package bootstrap

import (
	"context"
	cnst "duolingo/constant"
	config "duolingo/lib/config_reader"
	jr "duolingo/lib/config_reader/driver/reader/json"
	log "duolingo/lib/log"
	mq "duolingo/lib/message_queue"
	"duolingo/lib/message_queue/driver/rabbitmq"
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service_container"
	md "duolingo/service/message_input_api/server/middleware"
	"path/filepath"
	"time"
)

var (
	container *sv.ServiceContainer
	ctx       context.Context

	graceTimeOut   = 100 * time.Millisecond
	connTimeOut    = 10 * time.Second
	heartBeat      = 10 * time.Second
	declareTimeOut = 10 * time.Second
	writeTimeOut   = 5 * time.Second
)

func Run() {
	container = sv.GetContainer()
	ctx = context.Background()

	bindConfigReader()
	bindLogger()
	bindRestHttp()
	bindMessageQueue()
}

func bindConfigReader() {
	container.BindSingleton("config", func() any {
		dir, _ := filepath.Abs(filepath.Join(".", "config"))
		conf := jr.NewJsonReader(dir)
		return conf
	})
}

func bindLogger() {
	conf := container.Resolve("config").(config.ConfigReader)
	container.BindSingleton("log.server", func() any {
		dir, _ := filepath.Abs(".")
		rotation := time.Duration(conf.GetInt("message_input_api.log.rotation", 86400)) * time.Second
		flush := time.Duration(conf.GetInt("message_input_api.log.flush", 300)) * time.Second
		bufferSize := conf.GetInt("message_input_api.log.buffer.size", 2)
		bufferCount := conf.GetInt("message_input_api.log.buffer.max_count", 1000)

		return log.NewLoggerBuilder(ctx).
			UseNamespace("service", cnst.ServiceTypes[cnst.SV_INP_MESG], cnst.SV_INP_MESG).
			UseJsonFormat().
			AddLocalWriter(filepath.Join(dir, "service", cnst.SV_INP_MESG, "storage", "log")).
			WithFilePrefix(cnst.SV_INP_MESG).
			WithBuffering(bufferSize, bufferCount).
			WithRotation(rotation).
			WithFlushInterval(flush).
			Get()
	})
}

func bindRestHttp() {
	conf := container.Resolve("config").(config.ConfigReader)

	container.BindSingleton("rest.server", func() any {
		logger := container.Resolve("log.server").(*log.Logger)
		server := rest.
			NewServer(conf.Get("message_input_api.server.address", ":80")).
			WithMiddlewares("response", &md.LogHandledRequest{Logger: logger})

		return server
	})
}

func bindMessageQueue() {
	conf := container.Resolve("config").(config.ConfigReader)

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
				conf.Get("mq.host", ""),
				conf.Get("mq.port", ""),
				conf.Get("mq.user", ""),
				conf.Get("mq.pwd", ""),
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
