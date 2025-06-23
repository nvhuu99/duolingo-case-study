package bootstrap

import (
	"context"
	"fmt"

	"duolingo/constants"
	facade "duolingo/libraries/connection_manager/facade"
	"duolingo/libraries/pub_sub"
	"duolingo/libraries/pub_sub/drivers/rabbitmq"
	container "duolingo/libraries/service_container"
)

func BindPublisher() {
	container.Bind[pub_sub.Subscriber](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		subscriber := rabbitmq.NewRabbitMQSubscriber(provider.GetRabbitMQClient())
		topics := []string{
			constants.TopicPushNotiMessages,
		}
		for t := range topics {
			if subscribeErr := subscriber.Subscribe(topics[t]); subscribeErr != nil {
				panic(fmt.Errorf("failed to subscribe topic: %v", topics[t]))
			}
		}
		return subscriber
	})
}
