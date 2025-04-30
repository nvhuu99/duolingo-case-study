package bootstrap

import (
	"context"
	cnst "duolingo/constant"
	eh "duolingo/event/event_handler"
	config "duolingo/lib/config_reader"
	jr "duolingo/lib/config_reader/driver/reader/json"
	ep "duolingo/lib/event"
	log "duolingo/lib/log"
	mq "duolingo/lib/message_queue"
	"duolingo/lib/message_queue/driver/rabbitmq"
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service_container"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

var (
	container *sv.ServiceContainer
	ctx       context.Context
	cancel    context.CancelFunc

	graceTimeOut   = 100 * time.Millisecond
	connTimeOut    = 10 * time.Second
	heartBeat      = 10 * time.Second
	declareTimeOut = 10 * time.Second
	writeTimeOut   = 5 * time.Second
)

func Run() {
	container = sv.GetContainer()
	ctx, cancel = context.WithCancel(context.Background())

	bindContext()
	bindConfigReader()
	bindLogger()
	bindRestHttp()
	bindMessageQueue()
	bindEvents()
}

func bindContext() {
	container.BindSingleton("server.ctx", func() any { return ctx })
	container.BindSingleton("server.ctx_cancel", func() any { return cancel })
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
	container.BindSingleton("server.logger", func() any {
		dir, _ := filepath.Abs(".")
		rotation := time.Duration(conf.GetInt("input_message_api.log.rotation_seconds", 86400)) * time.Second
		flush := time.Duration(conf.GetInt("input_message_api.log.flush_seconds", 300)) * time.Second
		flushGrace := time.Duration(conf.GetInt("input_message_api.log.flush_grace_ms", 300)) * time.Millisecond
		bufferSize := conf.GetInt("input_message_api.log.buffer.size", 2)
		bufferCount := conf.GetInt("input_message_api.log.buffer.max_count", 1000)
		gRPCServerAddress := conf.Get("log_service.server.address", ":8002")

		uri := strings.Join([]string{"service", cnst.ServiceTypes[cnst.SV_INP_MESG], cnst.SV_INP_MESG}, "/")
		logger, err := log.NewLoggerBuilder(ctx).
			SetLogLevel(log.LevelAll).
			SetURI(uri).
			UseJsonFormat().
			WithBuffering(bufferSize, bufferCount).
			WithRotation(rotation).
			WithFlushInterval(flush, flushGrace).
			WithLocalFileOutput(filepath.Join(dir, "service", cnst.SV_INP_MESG, "storage", "log")).
			WithGRPCServiceOutput(gRPCServerAddress).
			Get()

		if err != nil {
			panic(err)
		}

		return logger
	})
}

func bindEvents() {
	evt := ep.NewEventPublisher()
	container.BindSingleton("event.publisher", func() any { return evt })
	evt.SubscribeRegex("service_operation_trace_.+", eh.NewSvOptTrace())
	evt.SubscribeRegex("service_operation_metric_.+", eh.NewSvOptMetric())
	evt.SubscribeRegex("input_message_request.+", eh.NewInputMessage())
}

func bindRestHttp() {
	conf := container.Resolve("config").(config.ConfigReader)

	container.BindSingleton("rest.server", func() any {
		server := rest.
			NewServer(fmt.Sprintf("0.0.0.0:%v", conf.Get("input_message_api.server.port", "8001")))
			// WithMiddlewares("request", new(md.RequestBegin)).
			// WithMiddlewares("response", new(md.RequestEnd))

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
