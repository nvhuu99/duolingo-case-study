package pub_sub

import (
	"context"
	connection "duolingo/libraries/connection_manager/drivers/rabbitmq"
	events "duolingo/libraries/events/facade"
	driver "duolingo/libraries/message_queue/drivers/rabbitmq"
	ps "duolingo/libraries/message_queue/pub_sub"
	"fmt"

	"github.com/google/uuid"
)

type Publisher struct {
	*driver.Publisher

	mainTopic string
}

func NewPublisher(client *connection.RabbitMQClient) *Publisher {
	return &Publisher{
		Publisher: &driver.Publisher{
			Topology: driver.NewTopology(client),
		},
	}
}

func (p *Publisher) SetMainTopic(topic string) {
	p.mainTopic = topic
}

func (p *Publisher) DeclareMainTopic() error {
	if p.mainTopic == "" {
		return ps.ErrPublisherMainTopicNotSet
	}
	return p.DeclareTopic(p.mainTopic)
}

func (p *Publisher) RemoveMainTopic() error {
	if p.mainTopic == "" {
		return ps.ErrPublisherMainTopicNotSet
	}
	topic := p.mainTopic
	p.mainTopic = ""
	return p.RemoveTopic(topic)
}

func (p *Publisher) NotifyMainTopic(ctx context.Context, message string) error {
	if p.mainTopic == "" {
		return ps.ErrPublisherMainTopicNotSet
	}
	return p.Notify(ctx, p.mainTopic, message)
}

func (p *Publisher) DeclareTopic(topic string) error {
	return p.DeclareExchange(
		driver.
			DefaultExchangeOpts(topic).
			IsType(driver.TopicExchange).
			IsPersistent(),
	)
}

func (p *Publisher) RemoveTopic(topic string) error {
	return p.DeleteExchange(topic)
}

func (p *Publisher) Notify(ctx context.Context, topic string, message string) error {
	var err error

	evtCtx, event := events.Start(ctx, fmt.Sprintf("pub_sub.publisher.notify(%v)", topic), nil)
	defer func() {
		events.End(event, err == nil, err, nil)
	}()

	err = p.Publish(evtCtx, topic, topic, message, map[string]string{
		"message_id": uuid.NewString(),
	})

	return err
}
