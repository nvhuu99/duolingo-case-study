package events

import (
	"context"
	"sync/atomic"
	"time"
)

var (
	eventManager           *EventManager
	eventManagerInitCalled atomic.Bool
)

type startEventRequest struct {
	event                 *Event
	eventTreeNodeTemplate *EventTreeNodeTemplate
	startedAt             time.Time
}

type endEventRequest struct {
	event   *Event
	endedAt time.Time
}

type EventManager struct {
	eventTree        *EventTreeRoot
	eventTreeBuilder *EventTreeBuilder

	ctx             context.Context
	opsChan         chan any
	collectInterval time.Duration

	subscribers []Subscriber
}

func Init(ctx context.Context, collectInterval time.Duration) {
	if eventManagerInitCalled.Load() {
		return
	}
	defer eventManagerInitCalled.Store(true)

	eventManager = &EventManager{
		ctx:              ctx,
		collectInterval:  collectInterval,
		opsChan:          make(chan any, 500),
		eventTree:        NewEventTreeRoot(),
		eventTreeBuilder: &EventTreeBuilder{},
	}

	go eventManager.handleOperationsChannel()

	go eventManager.collectEndedEventsAndNotifySubscribers()
}

func GetManager() *EventManager {
	return eventManager
}

func (m *EventManager) AddSubsriber(sub Subscriber) {
	m.subscribers = append(m.subscribers, sub)
}

func (m *EventManager) NewEvent(
	ctx context.Context,
	name string,
) (context.Context, *Event) {
	newEvent := NewEvent(name)
	newCtx, newNodeTemplate := m.eventTreeBuilder.NewNodeTemplate(ctx, newEvent)
	newEvent.ctx = newCtx

	m.opsChan <- &startEventRequest{
		event:                 newEvent,
		eventTreeNodeTemplate: newNodeTemplate,
		startedAt:             time.Now(),
	}

	// log.Println("queue create event for", name)

	return newCtx, newEvent
}

func (m *EventManager) EndEvent(event *Event, endedAt time.Time) {
	m.opsChan <- &endEventRequest{
		event: event, 
		endedAt: endedAt,
	}
	// log.Println("queue end event for", event.name)
}

func (m *EventManager) handleOperationsChannel() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case operation := <-m.opsChan:
			if op, start := operation.(*startEventRequest); start {
				m.startEvent(op)
			} else if op, end := operation.(*endEventRequest); end {
				m.endEvent(op)
			}
		}
	}
}

func (m *EventManager) startEvent(operation *startEventRequest) {
	// defer log.Println("created event map for", operation.event.name)

	event := operation.event
	event.startedAt = operation.startedAt

	eventNode := m.eventTree.builder.NewNode(operation.eventTreeNodeTemplate)
	m.eventTree.InsertNode(eventNode)

	go m.handleEventContextCancel(event)
}

func (m *EventManager) endEvent(op *endEventRequest) {
	event := op.event
	m.eventTree.FindNodeFromContextAndFlagEventEnded(event.ctx, op.endedAt)
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
		sub.Notified(event)
	}
}

func (m *EventManager) handleEventContextCancel(event *Event) {
	<- event.ctx.Done()

	if event.endedAt.IsZero() {
		m.EndEvent(event, time.Now())
	}
}