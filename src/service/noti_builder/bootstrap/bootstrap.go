package bootstrap

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	cnst "duolingo/constant"
	eh "duolingo/event/event_handler"
	config "duolingo/lib/config_reader"
	jr "duolingo/lib/config_reader/driver/reader/json"
	ep "duolingo/lib/event"
	"duolingo/lib/log"
	mq "duolingo/lib/message_queue"
	rabbitmq "duolingo/lib/message_queue/driver/rabbitmq"
	sv "duolingo/lib/service_container"
	distributor "duolingo/lib/work_distributor/driver/redis"
	db "duolingo/repository/campaign_db"
)

var (
	container *sv.ServiceContainer
	ctx       context.Context
	cancel    context.CancelFunc

	graceTimeOut   = 100 * time.Millisecond
	connTimeOut    = 10 * time.Second
	heartBeat      = 10 * time.Second
	declareTimeOut = 10 * time.Second
	writeTimeOut   = 10 * time.Second
)

func Run() {
	container = sv.GetContainer()
	ctx, cancel = context.WithCancel(context.Background())

	bindContext()
	bindConfigReader()
	bindLogger()
	bindRepository()
	bindWorkDistributor()
	bindMessageQueue()
	bindEvents()
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
		rotation := time.Duration(conf.GetInt("noti_builder.log.rotation_seconds", 86400)) * time.Second
		flush := time.Duration(conf.GetInt("noti_builder.log.flush_seconds", 300)) * time.Second
		flushGrace := time.Duration(conf.GetInt("noti_builder.log.flush_grace_ms", 300)) * time.Millisecond
		bufferSize := conf.GetInt("noti_builder.log.buffer.size", 2)
		bufferCount := conf.GetInt("noti_builder.log.buffer.max_count", 1000)
		gRPCServerAddress := conf.Get("log_service.server.address", ":8003")

		uri := strings.Join([]string{"service", cnst.ServiceTypes[cnst.SV_NOTI_BUILDER], cnst.SV_NOTI_BUILDER}, "/")
		logger, err := log.NewLoggerBuilder(ctx).
			SetLogLevel(log.LevelAll).
			SetURI(uri).
			UseJsonFormat().
			WithBuffering(bufferSize, bufferCount).
			WithRotation(rotation).
			WithFlushInterval(flush, flushGrace).
			WithLocalFileOutput(filepath.Join(dir, "service", cnst.SV_NOTI_BUILDER, "storage", "log")).
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
	evt.SubscribeRegex("relay_input_message_.+", eh.NewRelayInpMsg())
	evt.SubscribeRegex("build_push_notification_message.+", eh.NewBuildPushNotiMsg())
}

func bindRepository() {
	conf := container.Resolve("config").(config.ConfigReader)
	container.BindSingleton("repo.campaign_user", func() any {
		repo := db.NewUserRepo(ctx, conf.Get("db.campaign.name", ""))
		err := repo.SetConnection(
			conf.Get("db.campaign.host", ""),
			conf.Get("db.campaign.port", ""),
			conf.Get("db.campaign.user", ""),
			conf.Get("db.campaign.password", ""),
		)
		if err != nil {
			panic(err)
		}
		return repo
	})
}

func bindWorkDistributor() {
	conf := container.Resolve("config").(config.ConfigReader)
	container.BindSingleton("distributor", func() any {
		size := conf.GetInt("distributor.campaign_users.distribution_size", 1000)
		distributor, _ := distributor.NewRedisDistributor(ctx, "campaign_users")
		distributor.
			WithOptions(nil).
			WithLockTimeOut(100 * time.Second).
			WithDistributionSize(size)

		err := distributor.SetConnection(conf.Get("redis.host", ""), conf.Get("redis.port", ""))
		if err != nil {
			panic("failed to setup redis work distributor")
		}

		return distributor
	})
}

func bindMessageQueue() {
	conf := container.Resolve("config").(config.ConfigReader)

	container.BindSingleton("err_chan", func() any {
		return make(chan error, 10)
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
		errChan, _ := container.Resolve("err_chan").(chan error)

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
		topology.
			Topic("campaign_messages").Queue("push_noti_messages").Bind("push_noti_messages")

		return topology
	})

	container.BindSingleton("mq.publisher.input_messages", func() any {
		manager, _ := container.Resolve("mq.manager").(mq.Manager)
		errChan, _ := container.Resolve("err_chan").(chan error)

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

	container.BindSingleton("mq.publisher.push_noti_messages", func() any {
		manager, _ := container.Resolve("mq.manager").(mq.Manager)
		errChan, _ := container.Resolve("err_chan").(chan error)

		publisher := rabbitmq.
			NewPublisher("push_noti_messages_publisher", ctx)
		publisher.
			UseManager(manager)
		publisher.
			NotifyError(errChan)
		publisher.
			WithOptions(nil).
			WithGraceTimeOut(graceTimeOut).
			WithWriteTimeOut(writeTimeOut).
			WithTopic("campaign_messages").
			WithDirectDispatch("push_noti_messages")

		return publisher
	})

	container.BindSingleton("mq.consumer.input_messages", func() any {
		manager, _ := container.Resolve("mq.manager").(mq.Manager)
		errChan, _ := container.Resolve("err_chan").(chan error)

		consumer := rabbitmq.
			NewConsumer("input_messages_consumer", context.Background())
		consumer.
			UseManager(manager)
		consumer.
			NotifyError(errChan)
		consumer.
			WithOptions(nil).
			WithGraceTimeOut(graceTimeOut).
			WithQueue("input_messages")

		return consumer
	})
}
