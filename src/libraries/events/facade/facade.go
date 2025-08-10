package facade

import (
	"context"
	"duolingo/libraries/events"
	"sync"
	"time"
)

var (
	eventManager           *events.EventManager
	eventManagerInitOnce sync.Once
)

func InitEventManager(ctx context.Context, collectInterval time.Duration) {
	eventManagerInitOnce.Do(func() {
		eventManager = events.NewEventManager(ctx, collectInterval)
		eventManager.Start()
	})
}

func GetManager() *events.EventManager {
	return eventManager
}

func Subscribe(topicRegex string, sub events.Subscriber) {
	GetManager().AddSubsriber(topicRegex, sub)
}

func AddDecorators(decorators ...events.EventDecorator) {
	for _, d := range decorators {
		GetManager().AddDecorator(d)
	}
}


func AddFinalizer(finalizer events.EventFinalizer) {
	GetManager().AddFinalizer(finalizer)
}

func Start(ctx context.Context, name string, data map[string]any) *events.Event {
	return GetManager().StartEvent(ctx, name, data)
}

func End(event *events.Event, success bool, err error, data map[string]any) {
	if success && err == nil {
		Succeeded(event, data)
	} else {
		Failed(event, err, data)
	}
}

func Succeeded(event *events.Event, data map[string]any) {
	GetManager().EndEvent(event, time.Now(), events.EventSuccess, nil, data)
}

func Failed(event *events.Event, err error, data map[string]any) {
	GetManager().EndEvent(event, time.Now(), events.EventFailed, err, data)
}

