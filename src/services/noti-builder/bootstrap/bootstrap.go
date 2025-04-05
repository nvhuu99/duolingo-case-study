package bootstrap

import (
	"context"
	"duolingo/common"
	"time"

	cnst "duolingo/common/constant"

	config "duolingo/lib/config_reader"
	"duolingo/lib/log"
	mq "duolingo/lib/message-queue"
	rabbitmq "duolingo/lib/message-queue/driver/rabbitmq"
	sv "duolingo/lib/service-container"
	distributor "duolingo/lib/work-distributor/driver/redis"
	db "duolingo/repository/campaign-db"
)

var (
	container *sv.ServiceContainer
	ctx       context.Context
	infra     config.ConfigReader
	conf		config.ConfigReader

	graceTimeOut   = 100 * time.Millisecond
	connTimeOut    = 10 * time.Second
	heartBeat      = 10 * time.Second
	declareTimeOut = 10 * time.Second
	writeTimeOut   = 10 * time.Second
)

func Run() {
	common.SetupService()

	container = common.Container()
	ctx, _ = common.ServiceContext()
	infra, _ = container.Resolve("config.infra").(config.ConfigReader)
	conf		= container.Resolve("config").(config.ConfigReader)

	bindLogger()
	bindRepository()
	bindWorkDistributor()
	bindMessageQueue()
}

func bindLogger() {
	container.BindSingleton("log.server", func() any {
		rotation := time.Duration(conf.GetInt("self.log.rotation", 86400)) * time.Second
		flush := time.Duration(conf.GetInt("self.log.flush", 300)) * time.Second
		bufferSize := conf.GetInt("self.log.buffer.size", 2)
		bufferCount := conf.GetInt("self.log.buffer.max_count", 1000)

		return log.NewLoggerBuilder(ctx).
			UseNamespace("services", cnst.ServiceTypes[cnst.SV_NOTI_BUILDER], cnst.SV_NOTI_BUILDER).
			UseJsonFormat().
			AddLocalWriter(common.Dir("storage/log")).
			WithFilePrefix(cnst.SV_NOTI_BUILDER).
			WithBuffering(bufferSize, bufferCount).
			WithRotation(rotation).
			WithFlushInterval(flush).
			Get()
	})
}

func bindRepository() {
	container.BindSingleton("repo.campaign_user", func() any {
		repo := db.NewUserRepo(ctx, infra.Get("db.campaign.name", ""))
		repo.SetConnection(
			infra.Get("db.campaign.host", ""),
			infra.Get("db.campaign.port", ""),
			infra.Get("db.campaign.user", ""),
			infra.Get("db.campaign.password", ""),
		)
		return repo
	})
}

func bindWorkDistributor() {
	container.BindSingleton("distributor", func() any {
		size := infra.GetInt("distributor.campaign_users.distribution_size", 1000)
		distributor, _ := distributor.NewRedisDistributor(ctx, "campaign_users")
		distributor.
			WithOptions(nil).
			WithLockTimeOut(5 * time.Second).
			WithDistributionSize(size)

		err := distributor.SetConnection(infra.Get("redis.host", ""), infra.Get("redis.port", ""))
		if err != nil {
			panic("failed to setup redis work distributor")
		}

		return distributor
	})
}

func bindMessageQueue() {
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