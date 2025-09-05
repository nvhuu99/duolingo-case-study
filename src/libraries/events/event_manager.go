package events

import (
	"context"
	"regexp"
	"strings"
	"time"
)

type startEventRequest struct {
	event                 *Event
	eventTreeNodeTemplate *EventTreeNodeTemplate
}

type endEventRequest struct {
	event   *Event
	endedAt time.Time
	status  EventStatus
	err     error
}

type EventManager struct {
	ctx context.Context

	eventTree        *EventTreeRoot
	eventTreeBuilder *EventTreeBuilder
	opsChan          chan any
	collectInterval  time.Duration

	eventDecorators *EventPipeline
	eventFinalizers *EventPipeline

	subscribers      []Subscriber
	subscriberTopics map[string][]*regexp.Regexp // map subscriber id with pattern regexes
}

func NewEventManager(ctx context.Context, collectInterval time.Duration) *EventManager {
	return &EventManager{
		ctx:              ctx,
		eventTree:        NewEventTreeRoot(),
		eventTreeBuilder: &EventTreeBuilder{},
		collectInterval:  collectInterval,
		eventDecorators:  &EventPipeline{},
		eventFinalizers:  &EventPipeline{},
		opsChan:          make(chan any, 500),
		subscriberTopics: make(map[string][]*regexp.Regexp),
	}
}

func (m *EventManager) Start() {
	go m.handleOperationsChannel()
	go m.collectEndedEventsAndNotifySubscribers()
}

func (m *EventManager) AddDecorator(decorator EventDecorator) {
	m.eventDecorators.Push(NewBaseEventProcessor(decorator.Decorate))
}

func (m *EventManager) AddFinalizer(finalizer EventFinalizer) {
	m.eventFinalizers.Push(NewBaseEventProcessor(finalizer.Finalize))
}

func (m *EventManager) AddDecoratorFunc(prefix string, decoratorFunc func(*Event, *EventBuilder)) {
	m.eventDecorators.Push(NewBaseEventProcessor(func(e *Event, b *EventBuilder) {
		if strings.HasPrefix(e.Name(), prefix) {
			decoratorFunc(e, b)
		}
	}))
}

func (m *EventManager) AddSubscriber(pattern string, subscriber Subscriber) {
	id := subscriber.GetId()
	m.subscriberTopics[id] = append(m.subscriberTopics[id], regexp.MustCompile(pattern))
	m.subscribers = append(m.subscribers, subscriber)
}

func (m *EventManager) AddSubscribeFunc(pattern string, subscribeFunc func(*Event)) {
	subscriber := NewBaseEventSubscriber()
	subscriber.notifyFunc = subscribeFunc
	m.AddSubscriber(pattern, subscriber)
}

func (m *EventManager) StartEvent(
	ctx context.Context,
	name string,
	data map[string]any,
) *Event {
	newEvent := NewEvent(name, data)
	preparedCtx, newNodeTemplate := m.eventTreeBuilder.NewNodeTemplate(ctx, newEvent)
	newEvent.ctx = preparedCtx

	m.eventDecorators.Process(newEvent, NewEventBuilder(newEvent))

	m.opsChan <- &startEventRequest{
		event:                 newEvent,
		eventTreeNodeTemplate: newNodeTemplate,
	}

	return newEvent
}

func (m *EventManager) EndEvent(
	event *Event,
	endedAt time.Time,
	status EventStatus,
	err error,
	data map[string]any,
) {
	event.MergeData(data)

	m.eventFinalizers.Process(event, NewEventBuilder(event))

	m.opsChan <- &endEventRequest{
		event:   event,
		endedAt: endedAt,
		status:  status,
		err:     err,
	}
}

func (m *EventManager) handleOperationsChannel() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case operation := <-m.opsChan:
			if op, start := operation.(*startEventRequest); start {
				m.handleStartEventRequest(op)
			} else if op, end := operation.(*endEventRequest); end {
				m.handleEndEventRequest(op)
			}
		}
	}
}

func (m *EventManager) handleStartEventRequest(req *startEventRequest) {
	eventNode := m.eventTree.builder.NewNode(req.eventTreeNodeTemplate)
	m.eventTree.InsertNode(eventNode)
	m.notifySubscribers(req.event)
	go m.handleEventContextCancel(req.event)
}

func (m *EventManager) handleEndEventRequest(req *endEventRequest) {
	event := req.event
	event.status = req.status
	event.err = req.err
	m.eventTree.FindNodeFromContextAndFlagEventEnded(event.ctx, req.endedAt)
}

func (m *EventManager) collectEndedEventsAndNotifySubscribers() {
	ticker := time.NewTicker(m.collectInterval)
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			for _, event := range m.eventTree.ExtractAllEndedEvents() {
				m.notifySubscribers(event)
			}
		}
	}
}

func (m *EventManager) notifySubscribers(event *Event) {
	for _, sub := range m.subscribers {
		for _, pattern := range m.subscriberTopics[sub.GetId()] {
			if pattern.MatchString(event.name) {
				sub.Notify(event)
			}
		}
	}
}

func (m *EventManager) handleEventContextCancel(event *Event) {
	<-event.ctx.Done()
	// The context canceled, but endedAt has already set
	// means the event has ended normally
	if !event.endedAt.IsZero() {
		return
	}
	// End the event with interupted
	m.EndEvent(event, time.Now(), EventInterupted, nil, nil)
}
