package providers

import (
	"context"
	"duolingo/libraries/config_reader"
	"duolingo/libraries/connection_manager/drivers/mongodb"
	"duolingo/libraries/connection_manager/drivers/rabbitmq"
	"duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
	event "duolingo/libraries/events"
	events "duolingo/libraries/events/facade"
	"duolingo/libraries/telemetry/otel_wrapper/log"
)

type ConnectionsProvider struct {
}

func (provider *ConnectionsProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	provider.registerConnections(bootstrapCtx)
	provider.logsInstrumentation()
}

func (c *ConnectionsProvider) Shutdown(shutdownCtx context.Context) {
	facade.Provider().Shutdown()
}

func (provider *ConnectionsProvider) registerConnections(bootstrapCtx context.Context) {

	container.BindSingleton[*facade.ConnectionProvider](func(ctx context.Context) any {
		config := container.MustResolve[config_reader.ConfigReader]()

		facade.InitProvider(bootstrapCtx)

		return facade.Provider().
			InitRabbitMQ(rabbitmq.
				DefaultRabbitMQConnectionArgs().
				SetCredentials(
					config.Get("rabbitmq", "username"),
					config.Get("rabbitmq", "password"),
				).
				SetHost(config.Get("rabbitmq", "host")).
				SetPort(config.Get("rabbitmq", "port")),
			).
			InitMongo(mongodb.
				DefaultMongoConnectionArgs().
				SetCredentials(
					config.Get("mongodb", "username"),
					config.Get("mongodb", "password"),
				).
				SetHost(config.Get("mongodb", "host")).
				SetPort(config.Get("mongodb", "port")),
			).
			InitRedis(redis.
				DefaultRedisConnectionArgs().
				SetHost(config.Get("redis", "host")).
				SetPort(config.Get("redis", "port")),
			)
	})
}

func (provider *ConnectionsProvider) logsInstrumentation() {

	logger := container.MustResolve[*log.Logger]()

	events.SubscribeFunc("connection_manager", func(e *event.Event) {
		logger.Write(logger.
			Debug(e.GetDataStr("message")).
			Namespace("connection_manager").
			Err(e.Error()).
			Data(map[string]any{
				"connection_name": e.GetData("connection_name"),
			}),
		)
	})
}
