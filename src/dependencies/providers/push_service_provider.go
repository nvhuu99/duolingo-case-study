package providers

import (
	"context"
	"fmt"

	"duolingo/apps/push_sender/server/test/fakes"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	event "duolingo/libraries/events"
	events "duolingo/libraries/events/facade"
	push_noti "duolingo/libraries/push_notification"
	driver "duolingo/libraries/push_notification/drivers/firebase"
	"duolingo/libraries/telemetry/otel_wrapper/log"
	"duolingo/libraries/telemetry/otel_wrapper/trace"

	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type PushServiceProvider struct {
}

func (provider *PushServiceProvider) Bootstrap(bootstrapCtx context.Context, scope string) {

	tracer := container.MustResolve[*trace.TraceManager]()
	logger := container.MustResolve[*log.Logger]()

	/* Register Push Service */

	if scope == "test" {
		provider.registerFakePushServiceProvider()
	} else {
		provider.registerFirebasePushServiceProvider()
	}

	/* Tracing Instrumentation */

	tracer.Decorate("push_noti.push_service.send_multicast", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("push_noti.push_service.driver", "firebase"),
			attribute.String("push_noti.push_service.operation", "send_multicast"),
			attribute.String("push_noti.multicast.platforms", data.Get("platforms")),
			attribute.String("push_noti.multicast.response.devices_total", data.Get("devices_total")),
			attribute.String("push_noti.multicast.response.success_total", data.Get("success_total")),
			attribute.String("push_noti.multicast.response.failure_total", data.Get("failure_total")),
		)
	})

	/* Logs Instrumentation */

	events.SubscribeFunc("push_noti.push_service.send_multicast", func(e *event.Event) {
		logger.Write(logger.
			UnlessError(
				e.Error(), "push notifications failure ",
				log.LevelInfo, "push notifications sent",
			).
			Data(map[string]any{
				"driver":        "firebase",
				"operation":     "send_multicast",
				"platforms":     e.GetData("platforms"),
				"devices_total": e.GetData("devices_total"),
				"success_total": e.GetData("success_total"),
				"failure_total": e.GetData("failure_total"),
			}),
		)
	})
}

func (provider *PushServiceProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *PushServiceProvider) registerFakePushServiceProvider() {
	container.BindSingleton[push_noti.PushService](func(ctx context.Context) any {
		return fakes.NewFakePushService()
	})
}

func (provider *PushServiceProvider) registerFirebasePushServiceProvider() {
	container.BindSingleton[push_noti.PushService](func(ctx context.Context) any {
		config := container.MustResolve[config_reader.ConfigReader]()
		cred := config.Get("firebase", "credentials")

		factory, factoryErr := driver.NewFirebasePushNotiFactory(ctx, cred)
		if factoryErr != nil {
			panic(fmt.Errorf("failed to setup push service with error: %v ", factoryErr))
		}

		service, serviceErr := factory.CreatePushService()
		if serviceErr != nil {
			panic(fmt.Errorf("failed to setup push service with error: %v ", serviceErr))
		}

		return service
	})
}
