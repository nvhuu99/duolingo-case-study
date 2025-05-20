package event

import (
	"fmt"
	"sync"
)

type EventPublisher struct {
	subscribers map[string]Subscriber
	strTopics   map[string][]*strTopicSubscription
	regexTopics map[string][]*regexTopicSubscription
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		subscribers: make(map[string]Subscriber),
		strTopics:   make(map[string][]*strTopicSubscription),
		regexTopics: make(map[string][]*regexTopicSubscription),
	}
}

func (p *EventPublisher) Subscribe(wait bool, topic string, sub Subscriber) error {
	if topic == "" {
		return fmt.Errorf(ErrorMessages[ERR_EMPTY_PATTERN])
	}

	id := sub.SubscriberId()
	if id == "" {
		return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_ID_EMPTY])
	}

	if _, exists := p.subscribers[id]; !exists {
		p.subscribers[id] = sub
		p.strTopics[id] = []*strTopicSubscription{}
	}

	for _, strTopic := range p.strTopics[id] {
		if strTopic.topic == topic {
			return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_TOPIC_EXISTED], topic)
		}
	}

	p.strTopics[id] = append(p.strTopics[id], &strTopicSubscription{
		topic: topic,
		wait:  wait,
	})

	return nil
}

func (p *EventPublisher) SubscribeRegex(wait bool, regex string, sub Subscriber) error {
	regexPattern, err := newRegexPattern(regex)
	if err != nil {
		return err
	}

	if regexPattern.isEmptyPattern() {
		return fmt.Errorf(ErrorMessages[ERR_EMPTY_PATTERN])
	}

	id := sub.SubscriberId()
	if id == "" {
		return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_ID_EMPTY])
	}

	if _, exists := p.subscribers[id]; !exists {
		p.subscribers[id] = sub
		p.regexTopics[id] = []*regexTopicSubscription{}
	}

	for _, rt := range p.regexTopics[id] {
		if rt.regex.equal(regexPattern) {
			return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_TOPIC_EXISTED], regex)
		}
	}

	p.regexTopics[id] = append(p.regexTopics[id], &regexTopicSubscription{
		regex: regexPattern,
		wait:  wait,
	})

	return nil
}

func (p *EventPublisher) UnSubscribe(topic string, sub Subscriber) error {
	id := sub.SubscriberId()
	if _, exists := p.subscribers[id]; !exists {
		return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_NOT_EXIST])
	}

	strTopics := []*strTopicSubscription{}
	regexTopics := []*regexTopicSubscription{}
	for _, str := range p.strTopics[id] {
		if str.topic != topic {
			strTopics = append(strTopics, str)
		}
	}
	for _, reg := range p.regexTopics[id] {
		if !reg.regex.match(topic) {
			regexTopics = append(regexTopics, reg)
		}
	}

	if len(strTopics) == 0 && len(regexTopics) == 0 {
		delete(p.subscribers, id)
	} else {
		p.strTopics[id] = strTopics
		p.regexTopics[id] = regexTopics
	}

	return nil
}

func (p *EventPublisher) Notify(topic string, data any) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for id, subscriber := range p.subscribers {
		matches := false
		wait := false
		for _, st := range p.strTopics[id] {
			if st.topic == topic {
				matches = true
				wait = st.wait
				break
			}
		}
		if !matches {
			for _, rt := range p.regexTopics[id] {
				if rt.regex.match(topic) {
					matches = true
					wait = rt.wait
					break
				}
			}
		}
		if matches {
			if wait {
				wg.Add(1)
			}
			go func() {
				defer func() {
					if wait {
						wg.Done()
					}
				}()
				subscriber.Notified(topic, data)
			}()
		}
	}
}
