package providers

import (
	"context"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	events "duolingo/libraries/events/facade"
	"duolingo/services/trace_service"
	"time"
)

type EventManagerProvider struct {
}

func (provider *EventManagerProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	config := container.MustResolve[config_reader.ConfigReader]()
	collectInterval := config.GetInt("event_manager", "collect_interval_seconds")

	events.InitEventManager(
		bootstrapCtx,
		time.Duration(collectInterval)*time.Second,
	)

	if scope == "test" {
		return
	}

	events.AddDecorator(trace_service.NewRabbitMQContextPropagator())
	events.AddDecorator(trace_service.NewGlobalTracer())
	events.Subscribe(".*", trace_service.NewGlobalTracer())
}

func (provider *EventManagerProvider) Shutdown(shutdownCtx context.Context) {
}
