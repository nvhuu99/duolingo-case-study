package event

import (
	"fmt"
	"slices"
	"sync"
)

type EventPublisher struct {
	subscribers map[string]Subcriber
	strTopics   map[string][]string // subscriber topics mapped by subscriber ids
	regexTopics map[string][]*RegexPattern
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		subscribers: make(map[string]Subcriber),
		strTopics:   make(map[string][]string),
		regexTopics: make(map[string][]*RegexPattern),
	}
}

func (p *EventPublisher) Subscribe(topic string, sub Subcriber) error {
	if topic == "" {
		return fmt.Errorf(ErrorMessages[ERR_EMPTY_PATTERN])
	}

	id := sub.SubcriberId()
	if id == "" {
		return fmt.Errorf(ErrorMessages[ERR_SUBCRIBER_ID_EMPTY])
	}

	if _, exists := p.subscribers[id]; !exists {
		p.subscribers[id] = sub
		p.strTopics[id] = []string{}
	}

	if slices.Contains(p.strTopics[id], topic) {
		return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_TOPIC_EXISTED], topic)
	}

	p.strTopics[id] = append(p.strTopics[id], topic)

	return nil
}

func (p *EventPublisher) SubscribeRegex(regex string, sub Subcriber) error {
	regexPattern, err := NewRegexPattern(regex)
	if err != nil {
		return err
	}

	if regexPattern.IsEmptyPattern() {
		return fmt.Errorf(ErrorMessages[ERR_EMPTY_PATTERN])
	}

	id := sub.SubcriberId()
	if id == "" {
		return fmt.Errorf(ErrorMessages[ERR_SUBCRIBER_ID_EMPTY])
	}

	if _, exists := p.subscribers[id]; !exists {
		p.subscribers[id] = sub
		p.regexTopics[id] = []*RegexPattern{}
	}

	for _, rt := range p.regexTopics[id] {
		if rt.Equal(regexPattern) {
			return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_TOPIC_EXISTED], regex)
		}
	}

	p.regexTopics[id] = append(p.regexTopics[id], regexPattern)

	return nil
}

func (p *EventPublisher) UnSubcribe(topic Pattern, sub Subcriber) error {
	id := sub.SubcriberId()
	if _, exists := p.subscribers[id]; !exists {
		return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_NOT_EXIST])
	}

	delete(p.subscribers, id)
	delete(p.strTopics, id)
	delete(p.regexTopics, id)

	return nil
}

func (p *EventPublisher) Notify(wg *sync.WaitGroup, topic string, data any) {
	for id, subscriber := range p.subscribers {
		matches := slices.Contains(p.strTopics[id], topic) || slices.ContainsFunc(p.regexTopics[id],
			func(rt *RegexPattern) bool { return rt.Match(topic) },
		)
		if matches {
			if wg != nil {
				wg.Add(1)
			}
			go func() {
				defer func() {
					if wg != nil {
						wg.Done()
					}
				}()

				subscriber.Notified(topic, data)
			}()
		}
	}
}
