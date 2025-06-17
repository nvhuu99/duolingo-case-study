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
	container.BindSingleton[pub_sub.Publisher](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		publisher := rabbitmq.NewRabbitMQPublisher(provider.GetRabbitMQClient())
		topics := []string{
			constants.TopicMessageInputs,
			constants.TopicNotiBuilderJobs,
			constants.TopicPushNotiMessages,
		}
		for t := range topics {
			if declareErr := publisher.DeclareTopic(topics[t]); declareErr != nil {
				panic(fmt.Errorf("failed to declare topic: %v", topics[t]))
			}
		}
		return publisher
	})

	container.Bind[pub_sub.Subscriber](func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		subscriber := rabbitmq.NewRabbitMQSubscriber(provider.GetRabbitMQClient())
		topics := []string{
			constants.TopicMessageInputs,
			constants.TopicNotiBuilderJobs,
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
