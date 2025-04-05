package bootstrap

import (
	"context"
	"duolingo/common"

	cnst "duolingo/common/constant"
	config "duolingo/lib/config_reader"
	log "duolingo/lib/log"
	mq "duolingo/lib/message-queue"
	"duolingo/lib/message-queue/driver/rabbitmq"
	noti "duolingo/lib/notification/sender/firebase"
	sv "duolingo/lib/service-container"
	"os"
	"time"
)

var (
	container *sv.ServiceContainer
	ctx       context.Context
	infra     config.ConfigReader
	conf      config.ConfigReader

	graceTimeOut   = 100 * time.Millisecond
	connTimeOut    = 10 * time.Second
	heartBeat      = 10 * time.Second
	declareTimeOut = 10 * time.Second
)

func Run() {
	common.SetupService()

	container = common.Container()
	infra = container.Resolve("config.infra").(config.ConfigReader)
	conf = container.Resolve("config").(config.ConfigReader)
	ctx, _ = common.ServiceContext()

	bindLogger()
	bindMessageQueue()
}

func bindLogger() {
	container.BindSingleton("logger", func() any {
		rotation := time.Duration(conf.GetInt("self.log.rotation", 86400)) * time.Second
		flush := time.Duration(conf.GetInt("self.log.flush", 300)) * time.Second
		bufferSize := conf.GetInt("self.log.buffer.size", 1)
		bufferCount := conf.GetInt("self.log.buffer.max_count", 1000)

		return log.NewLoggerBuilder(ctx).
			UseNamespace("services", cnst.ServiceTypes[cnst.SV_PUSH_SENDER], cnst.SV_PUSH_SENDER).
			UseJsonFormat().
			AddLocalWriter(common.Dir("storage/log")).
			WithFilePrefix(cnst.SV_PUSH_SENDER).
			WithBuffering(bufferSize, bufferCount).
			WithRotation(rotation).
			WithFlushInterval(flush).
			Get()
	})
}

func bindMessageQueue() {
	container.BindSingleton("err_chan", func() any {
		return make(chan error, 2)
	})

	container.BindSingleton("noti.sender", func() any {
		credentialsPath := common.Dir("..", "..", "infra", "config", "firebase-service-account-key.json")
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
