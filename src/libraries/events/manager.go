package events

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

/* Event */

type Event struct {
	ctx context.Context
	name string
	startedAt time.Time
	endedAt time.Time
}

/* EventMap */

type EventMap struct {
	event *Event

	id string
	path string
	endedFlag bool
	childs map[string]*EventMap
	parent *EventMap
	isRoot bool
}

/* EventManager */

type EventContextValue string

const (
	CtxValEventMapPath EventContextValue = "event_manager.event_map_path"
)

var (
	eventManager *EventManager
	ensureSingleton sync.Once
)

type EventManager struct {
	eventMapRoot *EventMap
	eventMapMutex sync.RWMutex

	eventCreateChan chan *Event
	eventEndChan chan *Event

	subscribers []Subscriber
}

func GetManager() *EventManager {
	ensureSingleton.Do(func() {
		eventManager = &EventManager{
			eventCreateChan: make(chan *Event, 500),
			eventEndChan: make(chan *Event, 500),
			eventMapRoot: &EventMap{
				isRoot: true,
				childs: make(map[string]*EventMap),
			},
		}
		go eventManager.handleEvents(context.Background())
		go eventManager.collectEvents(context.Background())
	})
	return eventManager
}

func (m *EventManager) AddSubsriber(sub Subscriber) {
	m.subscribers = append(m.subscribers, sub)
}

func (m *EventManager) joinPath(paths ...string) string {
	filtered := []string{}
	for i := range paths {
		if len(paths[i]) > 0 {
			filtered = append(filtered, paths[i])
		}
	}
	return strings.Join(filtered, ".")
}

func (m *EventManager) NewEvent(
	ctx context.Context,
	name string,
) (context.Context, *Event) {
	newEventMapId := name // uuid.NewString()
	parentEventMapPath := m.extractEventMapPath(ctx)
	newEventMapPath := m.joinPath(parentEventMapPath, newEventMapId)

	newEvent := &Event{
		ctx: context.WithValue(ctx, CtxValEventMapPath, newEventMapPath), 
		name: name,
		startedAt: time.Now(),
	}
	
	m.eventCreateChan <- newEvent
	log.Println("queue event create for", name)

	return newEvent.ctx, newEvent
}

func (m *EventManager) EndEvent(event *Event, endedAt time.Time) {
	if event.endedAt.Before(endedAt) {
		event.endedAt = endedAt
	}
	m.eventEndChan <- event
	log.Println("queue event end for", event.name)
}

func (m *EventManager) handleEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-m.eventCreateChan:
			m.createEvent(event)
		case event := <-m.eventEndChan:
			m.travelUpAndEndEvents(event)
		}
	}
}

func (m *EventManager) collectEvents(ctx context.Context) {
	ticker := time.NewTicker(2*time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.eventMapMutex.Lock()
			for _, eventMap := range m.collectEndedEvents() {
				m.notifyAll(eventMap)
				delete(m.eventMapRoot.childs, eventMap.id)
			}
			m.eventMapMutex.Unlock()
		}
	}
}

func (m *EventManager) notifyAll(eventMap *EventMap) {
	for _, sub := range m.subscribers {
		sub.Notified(eventMap.event)
	} 
	for _, child := range eventMap.childs {
		m.notifyAll(child)
	}
}

func (m *EventManager) collectEndedEvents() []*EventMap {
	result := []*EventMap{}
	for i := range m.eventMapRoot.childs {
		if m.isEnded(m.eventMapRoot.childs[i]) {
			result = append(result, m.eventMapRoot.childs[i])
		}
	}
	return result
}

func (m *EventManager) isEnded(eventMap *EventMap) bool {
	for i := range eventMap.childs {
		ended := m.isEnded(eventMap.childs[i])
		if !ended {
			return false
		}
	}
	if ! eventMap.endedFlag {
		return false
	} 
	return eventMap.endedFlag
} 

func (m *EventManager) createEvent(event *Event) {
	m.eventMapMutex.Lock()
	defer m.eventMapMutex.Unlock()
	defer log.Println("created event map for", event.name)

	eventMapPath := event.ctx.Value(CtxValEventMapPath).(string)
	eventMapId := m.extractEventMapId(event.ctx)
	
	parentEventMapPath := m.extractParentEventMapPath(event.ctx)
	parentEventMap := m.find(parentEventMapPath)

	eventMap := &EventMap{
		event: event,
		id: eventMapId,
		path: eventMapPath,
		childs: make(map[string]*EventMap),	
	}

	if parentEventMap == nil {
		eventMap.parent = m.eventMapRoot
		m.eventMapRoot.childs[eventMapId] = eventMap
	} else {
		eventMap.parent = parentEventMap
		parentEventMap.childs[eventMapId] = eventMap
	}
}

func (m *EventManager) find(path string) *EventMap {
	parts := strings.Split(path, ".")
	travel := m.eventMapRoot
	for i := range parts {
		for id, node := range travel.childs {
			if parts[i] != id {
				continue
			}
			if i == len(parts) - 1 {
				return node	
			}
			travel = node
			break
		}
	}
	return nil
}

func (m *EventManager) travelUpAndEndEvents(event *Event) {
	m.eventMapMutex.Lock()
	defer m.eventMapMutex.Unlock()

	eventMapPath := event.ctx.Value(CtxValEventMapPath).(string)
	eventMap := m.find(eventMapPath)
	if eventMap == nil {
		m.EndEvent(event, event.endedAt)
		return
	}

	eventMap.endedFlag = true
	
	if len(eventMap.childs) > 0 {
		return
	}

	if eventMap.parent != m.eventMapRoot && eventMap.parent.endedFlag {
		m.EndEvent(eventMap.parent.event, event.endedAt)
	}
}

func (m *EventManager) extractEventMapPath(ctx context.Context) string {
	eventPath, _ := ctx.Value(CtxValEventMapPath).(string)
	return eventPath
}

func (m *EventManager) extractParentEventMapPath(ctx context.Context) string {
	eventPath, ok := ctx.Value(CtxValEventMapPath).(string)
	if !ok {
		return ""
	}
	parts := strings.Split(eventPath, ".")
	return strings.Join(parts[0:len(parts)-1], ".")
}

func (m *EventManager) extractEventMapId(ctx context.Context) string {
	eventPath, _ := ctx.Value(CtxValEventMapPath).(string)
	parts := strings.Split(eventPath, ".")
	return parts[len(parts)-1]
}
