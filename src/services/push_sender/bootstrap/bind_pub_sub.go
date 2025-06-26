package bootstrap

import (
	"context"
	"fmt"

	cnst "duolingo/constants"
	facade "duolingo/libraries/connection_manager/facade"
	"duolingo/libraries/pub_sub/drivers/rabbitmq"
	container "duolingo/libraries/service_container"
)

func BindPublisher() {
	container.BindSingletonAlias(cnst.PushNotiSubscriber, func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		subscriber := rabbitmq.NewRabbitMQSubscriber(provider.GetRabbitMQClient())
		subscriber.SetMainTopic(cnst.PushNotiTopic)
		if subcribeErr := subscriber.SubscribeMainTopic(); subcribeErr != nil {
			panic(fmt.Errorf("failed to subscribe topic: %v", cnst.PushNotiTopic))
		}
		return subscriber
	})
}
