package providers

import (
	"context"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	"duolingo/libraries/telemetry/otel_wrapper/trace"
	"os"

	"go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/trace"
)

type TraceManagerProvider struct {
}

func (provider *TraceManagerProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	config := container.MustResolve[config_reader.ConfigReader]()

	appName := os.Getenv("DUOLINGO_APP_NAME")
	endpoint := config.Get("alloy", "otlp.receiver.grpc.endpoint")
	trace.InitTraceManager(
		bootstrapCtx,
		trace.WithDefaultResource(appName),
		trace.WithGRPCExporter(endpoint, false),
	)

	trace.GetManager().Decorate("restful.<method>(<path>)", func(
		span otlptrace.Span,
		data trace.DataBag,
	) {
		span.SetAttributes(
			attribute.String("kind", "server"),
			attribute.String("http.request.method", data.Get("method")),
			attribute.String("url.scheme", data.Get("scheme")),
			attribute.String("url.path", data.Get("path")),
			attribute.String("url.full", data.Get("full_url")),
			attribute.String("http.response.status_code", data.Get("status_code")),
			attribute.String("user_agent.original", data.Get("user_agent")),
		)
	})
}

func (provider *TraceManagerProvider) Shutdown(shutdownCtx context.Context) {
}
