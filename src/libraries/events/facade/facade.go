package facade

import (
	"context"
	"duolingo/libraries/events"
	"sync"
	"time"
)

var (
	eventManager         *events.EventManager
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
	GetManager().AddSubscriber(topicRegex, sub)
}

func SubscribeFunc(topicRegex string, subscribeFunc func(*events.Event)) {
	GetManager().AddSubscribeFunc(topicRegex, subscribeFunc)
}

func AddDecorator(decorator events.EventDecorator) {
	GetManager().AddDecorator(decorator)
}

func AddFinalizer(finalizer events.EventFinalizer) {
	GetManager().AddFinalizer(finalizer)
}

func AddDecoratorFunc(prefix string, decoratorFunc func(*events.Event, *events.EventBuilder)) {
	GetManager().AddDecoratorFunc(prefix, decoratorFunc)
}

func Emit(ctx context.Context, name string, err error, data map[string]any) {
	evt := GetManager().StartEvent(ctx, name, data)
	End(evt, true, err, data)
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
