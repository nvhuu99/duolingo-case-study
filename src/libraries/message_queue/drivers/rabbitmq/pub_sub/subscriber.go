package pub_sub

import (
	"context"
	"fmt"

	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	events "duolingo/libraries/events/facade"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq"
	ps "duolingo/libraries/message_queue/pub_sub"

	"github.com/google/uuid"
)

type Subscriber struct {
	*driver.QueueConsumer

	id        string
	queues    map[string]string // unique queues created for each topic subscribed
	mainTopic string
}

func NewSubscriber(client *connection.RabbitMQClient) *Subscriber {
	return &Subscriber{
		id:     uuid.NewString(),
		queues: make(map[string]string),
		QueueConsumer: &driver.QueueConsumer{
			Topology: driver.NewTopology(client),
		},
	}
}

func (sub *Subscriber) SetMainTopic(topic string) {
	sub.mainTopic = topic
}

func (sub *Subscriber) SubscribeMainTopic(ctx context.Context) error {
	if sub.mainTopic == "" {
		return ps.ErrSubscriberMainTopicNotSet
	}
	return sub.Subscribe(ctx, sub.mainTopic)
}

func (sub *Subscriber) UnSubscribeMainTopic(ctx context.Context) error {
	if sub.mainTopic == "" {
		return ps.ErrSubscriberMainTopicNotSet
	}
	topic := sub.mainTopic
	sub.mainTopic = ""
	return sub.UnSubscribe(ctx, topic)
}

func (sub *Subscriber) ListeningMainTopic(
	ctx context.Context,
	processFunc func(context.Context, string) error,
) error {
	if sub.mainTopic == "" {
		return ps.ErrSubscriberMainTopicNotSet
	}
	return sub.Listening(ctx, sub.mainTopic, processFunc)
}

func (sub *Subscriber) Subscribe(ctx context.Context, topic string) error {
	if _, exist := sub.queues[topic]; !exist {
		sub.queues[topic] = fmt.Sprintf("%v_%v", topic, sub.id)
	}
	return sub.bindQueue(ctx, topic)
}

func (sub *Subscriber) UnSubscribe(ctx context.Context, topic string) error {
	delete(sub.queues, topic)
	return nil
}

func (sub *Subscriber) Listening(
	ctx context.Context,
	topic string,
	processFunc func(context.Context, string) error,
) error {
	if _, exists := sub.queues[topic]; !exists {
		return ps.ErrSubscriberTopicNotSubscribed
	}

	if bindErr := sub.bindQueue(ctx, topic); bindErr != nil {
		return bindErr
	}

	return sub.Consuming(ctx, sub.queues[topic], func(
		ctx context.Context,
		msg string,
	) (driver.ConsumeAction, error) {
		var err error

		evt := events.Start(ctx, fmt.Sprintf("pub_sub.subscriber.notified(%v)", topic), map[string]any{
			"topic": topic,
		})
		defer events.End(evt, true, err, nil)

		err = processFunc(ctx, msg)

		return driver.ActionAccept, err
	})
}

func (sub *Subscriber) bindQueue(ctx context.Context, topic string) error {
	if _, exist := sub.queues[topic]; !exist {
		return ps.ErrSubscriberTopicNotSubscribed
	}

	var declareErr error
	declareErr = sub.DeclareExchange(
		ctx,
		driver.
			DefaultExchangeOpts(topic).
			IsType(driver.TopicExchange).
			IsPersistent(),
	)
	if declareErr == nil {
		declareErr = sub.DeclareQueue(
			ctx,
			driver.DefaultQueueOpts(sub.queues[topic]).
				IsNonPersistent().
				IsExclusive(),
			driver.NewQueueBinding(sub.queues[topic]).
				Add(topic, topic),
		)
	}

	return declareErr
}
