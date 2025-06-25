package bootstrap

import (
	"context"
	"fmt"

	facade "duolingo/libraries/connection_manager/facade"
	"duolingo/libraries/pub_sub/drivers/rabbitmq"
	container "duolingo/libraries/service_container"
	cnst "duolingo/constants"
)

func BindPublisher() {
	bindAndDeclare(cnst.MesgInputPublisher, cnst.MesgInputTopic)
	bindAndSubscribe(cnst.MesgInputSubscriber, cnst.MesgInputTopic)

	bindAndDeclare(cnst.NotiBuilderJobPublisher, cnst.NotiBuilderJobTopic)
	bindAndSubscribe(cnst.NotiBuilderJobSubscriber, cnst.NotiBuilderJobTopic)

	bindAndDeclare(cnst.PushNotiPublisher, cnst.PushNotiTopic)
	bindAndSubscribe(cnst.PushNotiSubscriber, cnst.PushNotiTopic)
}

func bindAndDeclare(publisher string, topic string) {
	container.BindSingletonAlias(publisher, func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		publisher := rabbitmq.NewRabbitMQPublisher(provider.GetRabbitMQClient())
		publisher.SetMainTopic(topic)
		if declareErr := publisher.DeclareMainTopic(); declareErr != nil {
			panic(fmt.Errorf("failed to declare topic: %v", cnst.MesgInputTopic))
		}
		return publisher
	})
}

func bindAndSubscribe(subscriber string, topic string) {
	container.BindSingletonAlias(subscriber, func(ctx context.Context) any {
		provider := container.MustResolve[*facade.ConnectionProvider]()
		subscriber := rabbitmq.NewRabbitMQSubscriber(provider.GetRabbitMQClient())
		subscriber.SetMainTopic(topic)
		if subscribeErr := subscriber.SubscribeMainTopic(); subscribeErr != nil {
			panic(fmt.Errorf("failed to subscribe topic: %v", topic))
		}
		return subscriber
	})
}