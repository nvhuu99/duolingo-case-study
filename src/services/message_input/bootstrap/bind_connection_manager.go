package bootstrap

import (
	"context"
	"duolingo/libraries/connection_manager/drivers/rabbitmq"
	"duolingo/libraries/connection_manager/facade"
	container "duolingo/libraries/service_container"
)

func BindConnections() {
	container.BindSingleton[*facade.ConnectionProvider](func(ctx context.Context) any {
		return facade.Provider(ctx).
			InitRabbitMQ(rabbitmq.
				DefaultRabbitMQConnectionArgs().
				SetCredentials("root", "12345"),
			)
	})

	container.Bind[*rabbitmq.RabbitMQClient](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		return provider.GetRabbitMQClient()
	})
}
