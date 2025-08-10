package dependencies

import (
	"context"
	"duolingo/libraries/config_reader"
	"duolingo/libraries/connection_manager/drivers/mongodb"
	"duolingo/libraries/connection_manager/drivers/rabbitmq"
	"duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/dependencies_container"
)

type Connections struct {
}

func NewConnections() *Connections {
	return &Connections{}
}

func (provider *Connections) Bootstrap(bootstrapCtx context.Context, scope string) {
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

func (c *Connections) Shutdown(shutdownCtx context.Context) {
	facade.Provider().Shutdown()
}
