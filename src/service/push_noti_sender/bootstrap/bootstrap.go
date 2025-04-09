package bootstrap

import (
	"context"
	"path/filepath"

	cnst "duolingo/constant"
	config "duolingo/lib/config_reader"
	jr "duolingo/lib/config_reader/driver/reader/json"
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

	graceTimeOut   = 100 * time.Millisecond
	connTimeOut    = 10 * time.Second
	heartBeat      = 10 * time.Second
	declareTimeOut = 10 * time.Second
)

func Run() {
	container = sv.GetContainer()
	ctx = context.Background()

	bindConfigReader()
	bindLogger()
	bindMessageQueue()
}

func bindConfigReader() {
	container.BindSingleton("config", func() any {
		conf := jr.NewJsonReader(filepath.Join(".", "config"))
		return conf
	})
}

func bindLogger() {
	conf := container.Resolve("config").(config.ConfigReader)
	container.BindSingleton("logger", func() any {
		dir, _ := filepath.Abs(".")
		rotation := time.Duration(conf.GetInt("push_noti_sender.log.rotation", 86400)) * time.Second
		flush := time.Duration(conf.GetInt("push_noti_sender.log.flush", 300)) * time.Second
		bufferSize := conf.GetInt("push_noti_sender.log.buffer.size", 1)
		bufferCount := conf.GetInt("push_noti_sender.log.buffer.max_count", 1000)

		return log.NewLoggerBuilder(ctx).
			UseNamespace("service", cnst.ServiceTypes[cnst.SV_PUSH_SENDER], cnst.SV_PUSH_SENDER).
			UseJsonFormat().
			AddLocalWriter(filepath.Join(dir, "service", cnst.SV_PUSH_SENDER, "storage", "log")).
			WithFilePrefix(cnst.SV_PUSH_SENDER).
			WithBuffering(bufferSize, bufferCount).
			WithRotation(rotation).
			WithFlushInterval(flush).
			Get()
	})
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
