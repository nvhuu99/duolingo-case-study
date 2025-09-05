package log

import (
	"context"
	"time"
)

type LoggerBuilder struct {
	ctx     context.Context
	level   LogLevel
	writers []LogWriter
}

func NewLoggerBuilder(ctx context.Context) *LoggerBuilder {
	return &LoggerBuilder{
		ctx: ctx,
	}
}

func (builder *LoggerBuilder) SetLogLevel(level LogLevel) *LoggerBuilder {
	builder.level = level
	return builder
}

func (builder *LoggerBuilder) WithConsoleOutput(formatter LogFormatter) *LoggerBuilder {
	builder.writers = append(builder.writers, NewConsoleWriter().WithFormatter(formatter))
	return builder
}

func (builder *LoggerBuilder) WithGrafanaLokiOutput(
	formatter LogFormatter,
	serviceName string,
	lokiEndpoint string,
	limit int,
	interval time.Duration,
) *LoggerBuilder {
	loki := NewLokiWriter(
		builder.ctx,
		serviceName,
		lokiEndpoint,
		limit,
		interval,
	)
	loki.WithFormatter(formatter)
	builder.writers = append(builder.writers, loki)
	return builder
}

func (builder *LoggerBuilder) GetLogger() *Logger {
	return &Logger{
		level:   builder.level,
		writers: builder.writers,
	}
}
