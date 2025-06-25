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
	container.BindSingletonAlias(cnst.MesgInputPublisher, func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		publisher := rabbitmq.NewRabbitMQPublisher(provider.GetRabbitMQClient())
		publisher.SetMainTopic(cnst.MesgInputTopic)
		if declareErr := publisher.DeclareMainTopic(); declareErr != nil {
			panic(fmt.Errorf("failed to declare topic: %v", cnst.MesgInputTopic))
		}
		return publisher
	})
}
