package providers

import (
	"context"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	"duolingo/libraries/telemetry/otel_wrapper/log"
	"os"
	"time"
)

type LoggerProvider struct {
}

func (provider *LoggerProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	config := container.MustResolve[config_reader.ConfigReader]()
	appName := os.Getenv("DUOLINGO_APP_NAME")
	endpoint := config.Get("instrumentation", "loki.endpoint")
	limit := config.GetInt("instrumentation", "log.buffer_limit_count")
	interval := config.GetInt("instrumentation", "log.buffer_flush_interval_seconds")
	level := config.Get("instrumentation", "log.level")

	container.BindSingleton[*log.Logger](func(ctx context.Context) any {
		return log.NewLoggerBuilder(ctx).
			SetLogLevel(log.ParseLogLevelString(level)).
			UseJsonFormat().
			WithConsoleOutput().
			WithGrafanaLokiOutput(
				appName,
				endpoint,
				limit,
				time.Duration(interval)*time.Second,
			).
			GetLogger()
	})
}

func (provider *LoggerProvider) Shutdown(shutdownCtx context.Context) {
}
