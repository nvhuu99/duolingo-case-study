package bootstrap

import (
	"fmt"
	"context"
	
	"duolingo/constants"
	facade "duolingo/libraries/connection_manager/facade"
	"duolingo/libraries/pub_sub"
	"duolingo/libraries/pub_sub/drivers/rabbitmq"
	container "duolingo/libraries/service_container"
)

func BindPublisher() {
	container.BindSingleton[pub_sub.Publisher](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		publisher := rabbitmq.NewRabbitMQPublisher(provider.GetRabbitMQClient())
		
		declareErr := publisher.DeclareTopic(constants.PubSubTopicMessageInput)
		if declareErr != nil {
			panic(fmt.Errorf("failed to declare topic: %v", constants.PubSubTopicMessageInput))
		} 
		
		return publisher
	})
}
