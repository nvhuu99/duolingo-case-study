package providers

import (
	"context"
	"duolingo/libraries/connection_manager/facade"
	"duolingo/repositories/user_repository/drivers/mongodb"
	user_repo "duolingo/repositories/user_repository/external"

	"duolingo/libraries/telemetry/otel_wrapper/log"
	"duolingo/libraries/telemetry/otel_wrapper/trace"

	container "duolingo/libraries/dependencies_container"
	event "duolingo/libraries/events"
	events "duolingo/libraries/events/facade"

	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type UserRepoProvider struct {
}

func (provider *UserRepoProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *UserRepoProvider) Bootstrap(bootstrapCtx context.Context, scope string) {

	tracer := container.MustResolve[*trace.TraceManager]()
	logger := container.MustResolve[*log.Logger]()

	/* Register Repository */

	provider.registerMongoDBUserRepo()

	/* Tracing Instrumentation */

	tracer.Decorate("user_repo.*", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("database.system.name", "mongodb"),
			attribute.String("database.collection.name", "users"),
			attribute.String("db.operation.name", data.Get("db_operation")),
			attribute.String("user_repo.operation.name", data.Get("operation_name")),
		)
	})

	/* Logs Instrumentation */

	events.SubscribeFunc("user_repo.*", func(e *event.Event) {
		logger.Write(logger.
			UnlessError(
				e.Error(), "operation success",
				log.LevelInfo, "operation failure",
			).
			Data(map[string]any{
				"database.system.name":     "mongodb",
				"database.collection.name": "users",
				"db.operation.name":        e.GetData("db_operation"),
				"user_repo.operation.name": e.GetData("operation_name"),
			}),
		)
	})
}

func (provider *UserRepoProvider) registerMongoDBUserRepo() {
	container.BindSingleton[user_repo.UserRepoFactory](func(ctx context.Context) any {
		connections := container.MustResolve[*facade.ConnectionProvider]()
		return mongodb.NewUserRepoFactory(connections.GetMongoClient())
	})

	container.BindSingleton[user_repo.UserRepository](func(ctx context.Context) any {
		factory := container.MustResolve[user_repo.UserRepoFactory]()
		return factory.MakeUserRepo()
	})
}
