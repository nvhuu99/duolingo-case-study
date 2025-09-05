package providers

import (
	"context"

	container "duolingo/libraries/dependencies_container"
	event "duolingo/libraries/events"
	events "duolingo/libraries/events/facade"
	"duolingo/libraries/telemetry/otel_wrapper/log"
	"duolingo/libraries/telemetry/otel_wrapper/trace"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/services/user_service"

	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type UserServiceProvider struct {
}

func (provider *UserServiceProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *UserServiceProvider) Bootstrap(bootstrapCtx context.Context, scope string) {

	tracer := container.MustResolve[*trace.TraceManager]()
	logger := container.MustResolve[*log.Logger]()

	/* Register User Service */

	provider.registerMongoDBUserService()

	/* Tracing Instrumentation */

	tracer.Decorate("user_service.*", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("user_service.operation.name", data.Get("operation_name")),
		)
	})

	/* Logs Instrumentation */

	events.SubscribeFunc("user_service.*", func(e *event.Event) {
		logger.Write(logger.
			UnlessError(
				e.Error(), "operation failure",
				log.LevelInfo, "operation success",
			).
			Data(map[string]any{
				"user_service.operation.name": e.GetData("operation_name"),
			}),
		)
	})
}

func (provider *UserServiceProvider) registerMongoDBUserService() {
	container.BindSingleton[*user_service.UserService](func(ctx context.Context) any {
		return user_service.NewUserService(
			container.MustResolve[user_repo.UserRepoFactory](),
			container.MustResolve[user_repo.UserRepository](),
		)
	})
}
