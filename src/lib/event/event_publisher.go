package event

import (
	"fmt"
	"slices"
)

type EventPublisher struct {
	listeners   map[string]Subcriber
	strTopics   map[string][]string // listener topics mapped by listener ids
	regexTopics map[string][]*RegexPattern
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		listeners:   make(map[string]Subcriber),
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

	if _, exists := p.listeners[id]; !exists {
		p.listeners[id] = sub
		p.strTopics[id] = []string{}
	}

	for _, t := range p.strTopics[id] {
		if t == topic {
			return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_TOPIC_EXISTED], topic)
		}
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

	if _, exists := p.listeners[id]; !exists {
		p.listeners[id] = sub
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
	if _, exists := p.listeners[id]; !exists {
		return fmt.Errorf(ErrorMessages[ERR_SUBSCRIBER_NOT_EXIST])
	}

	delete(p.listeners, id)
	delete(p.strTopics, id)
	delete(p.regexTopics, id)

	return nil
}

func (p *EventPublisher) Notify(topic string, data any) {
	for id, listener := range p.listeners {
		// match string
		if slices.Contains(p.strTopics[id], topic) {
			listener.Notified(topic, data)
			continue
		}
		// matching regex
		matches := slices.ContainsFunc(p.regexTopics[id], func(rt *RegexPattern) bool {
			return rt.Match(topic)
		})
		if matches {
			listener.Notified(topic, data)
		}
	}
}
