package bootstrap

import (
	"context"
	"path/filepath"
	"strings"

	cnst "duolingo/constant"
	eh "duolingo/event/event_handler"
	config "duolingo/lib/config_reader"
	jr "duolingo/lib/config_reader/driver/reader/json"
	ep "duolingo/lib/event"
	log "duolingo/lib/log"
	mq "duolingo/lib/message_queue"
	"duolingo/lib/message_queue/driver/rabbitmq"
	noti "duolingo/lib/notification/sender/firebase"
	sv "duolingo/lib/service_container"
	"os"
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
)

func Run() {
	container = sv.GetContainer()
	ctx = context.Background()
	ctx, cancel = context.WithCancel(context.Background())

	bindContext()
	bindConfigReader()
	bindLogger()
	bindEvents()
	bindMessageQueue()
}

func bindContext() {
	container.BindSingleton("server.ctx", func() any { return ctx })
	container.BindSingleton("server.ctx_cancel", func() any { return cancel })
}

func bindConfigReader() {
	container.BindSingleton("config", func() any {
		conf := jr.NewJsonReader(filepath.Join(".", "config"))
		return conf
	})
}

func bindLogger() {
	conf := container.Resolve("config").(config.ConfigReader)
	container.BindSingleton("server.logger", func() any {
		dir, _ := filepath.Abs(".")
		rotation := time.Duration(conf.GetInt("push_noti_sender.log.rotation_seconds", 86400)) * time.Second
		flush := time.Duration(conf.GetInt("push_noti_sender.log.flush_seconds", 300)) * time.Second
		flushGrace := time.Duration(conf.GetInt("push_noti_sender.log.flush_grace_ms", 300)) * time.Millisecond
		bufferSize := conf.GetInt("push_noti_sender.log.buffer.size", 1)
		bufferCount := conf.GetInt("push_noti_sender.log.buffer.max_count", 1000)

		uri := strings.Join([]string{ "service", cnst.ServiceTypes[cnst.SV_PUSH_SENDER], cnst.SV_PUSH_SENDER }, "/")
		return log.NewLoggerBuilder(ctx).
			SetLogLevel(log.LevelAll).
			SetURI(uri).
			UseJsonFormat().
			WithBuffering(bufferSize, bufferCount).
			WithRotation(rotation).
			WithFlushInterval(flush, flushGrace).
			WithLocalFileOutput(filepath.Join(dir, "service", cnst.SV_PUSH_SENDER, "storage", "log")).
			Get()
	})
}

func bindEvents() {
	evt := ep.NewEventPublisher()
	container.BindSingleton("event.publisher", func() any { return evt })
	evt.SubscribeRegex("service_operation_trace_.+", eh.NewSvOptTrace())
	evt.SubscribeRegex("service_operation_metric_.+", eh.NewSvOptMetric())
	evt.SubscribeRegex("send_push_notification_.+", eh.NewSendPushNoti())
}

func bindMessageQueue() {
	conf := container.Resolve("config").(config.ConfigReader)

	container.BindSingleton("err_chan", func() any {
		return make(chan error, 2)
	})

	container.BindSingleton("noti.sender", func() any {
		credentialsPath := filepath.Join("config", "firebase_service_account_key.json")
		credentials, err := os.ReadFile(credentialsPath)
		if err != nil {
			panic(err)
		}
		sender := noti.NewFirebaseSender(context.Background())
		err = sender.WithJsonCredentials(string(credentials))
		if err != nil {
			panic(err)
		}

		return sender
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
			Topic("campaign_messages").Queue("push_noti_messages").Bind("push_noti_messages")

		return topology
	})

	container.BindSingleton("mq.consumer", func() any {
		manager, _ := container.Resolve("mq.manager").(mq.Manager)
		errChan, _ := container.Resolve("err_chan").(chan error)

		consumer := rabbitmq.
			NewConsumer("push_noti_messages_consumer", context.Background())
		consumer.
			UseManager(manager)
		consumer.
			NotifyError(errChan)
		consumer.
			WithOptions(nil).
			WithGraceTimeOut(graceTimeOut).
			WithQueue("push_noti_messages")

		return consumer
	})
}
