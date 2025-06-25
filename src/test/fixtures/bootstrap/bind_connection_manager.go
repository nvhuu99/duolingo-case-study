package bootstrap

import (
	"context"
	"duolingo/libraries/connection_manager/drivers/mongodb"
	"duolingo/libraries/connection_manager/drivers/rabbitmq"
	"duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/service_container"
)

func BindConnections() {
	container.BindSingleton[*facade.ConnectionProvider](func(ctx context.Context) any {
		return facade.Provider(ctx).
			InitRabbitMQ(rabbitmq.
				DefaultRabbitMQConnectionArgs().
				SetCredentials("root", "12345"),
			).
			InitMongo(mongodb.
				DefaultMongoConnectionArgs().
				SetCredentials("root", "12345"),
			).
			InitRedis(redis.
				DefaultRedisConnectionArgs(),
			)
	})

	container.Bind[*redis.RedisClient](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return provider.GetRedisClient()
	})

	container.Bind[*mongodb.MongoClient](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return provider.GetMongoClient()
	})

	container.Bind[*rabbitmq.RabbitMQClient](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return provider.GetRabbitMQClient()
	})
}
