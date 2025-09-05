package providers

import (
	"context"

	"duolingo/libraries/config_reader"
	facade "duolingo/libraries/connection_manager/facade"
	"duolingo/libraries/telemetry/otel_wrapper/log"
	"duolingo/libraries/telemetry/otel_wrapper/trace"
	dist "duolingo/libraries/work_distributor"
	redis "duolingo/libraries/work_distributor/drivers/redis"

	container "duolingo/libraries/dependencies_container"
	event "duolingo/libraries/events"
	events "duolingo/libraries/events/facade"

	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type WorkDistributorProvider struct {
}

func (provider *WorkDistributorProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	tracer := container.MustResolve[*trace.TraceManager]()
	logger := container.MustResolve[*log.Logger]()

	/* Register Work Distributor */

	provider.registerRedisWorkDistributor()

	/* Tracing Instrumentation */

	tracer.Decorate("work_dist.*", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("work_distributor.operation.name", data.Get("operation_name")),
		)
	})

	/* Logs Instrumentation */

	events.SubscribeFunc("work_dist.*", func(e *event.Event) {
		logger.Write(logger.
			UnlessError(
				e.Error(), "operation failure",
				log.LevelInfo, "operation success",
			).
			Data(map[string]any{
				"work_distributor.operation.name": e.GetData("operation_name"),
			}),
		)
	})
}

func (provider *WorkDistributorProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *WorkDistributorProvider) registerRedisWorkDistributor() {
	container.BindSingleton[*dist.WorkDistributor](func(ctx context.Context) any {
		config := container.MustResolve[config_reader.ConfigReader]()
		connections := container.MustResolve[*facade.ConnectionProvider]()
		return redis.NewRedisWorkDistributor(
			connections.GetRedisClient(),
			config.GetInt64("work_distributor", "distribution_size"),
		)
	})
}
