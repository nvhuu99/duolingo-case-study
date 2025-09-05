package log

import (
	"context"
	"time"
)

type LoggerBuilder struct {
	ctx       context.Context
	level     LogLevel
	formatter LogFormatter
	writers   []LogWriter
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

func (builder *LoggerBuilder) UseJsonFormat() *LoggerBuilder {
	builder.formatter = &JsonFormatter{}
	return builder
}

func (builder *LoggerBuilder) WithConsoleOutput() *LoggerBuilder {
	builder.writers = append(builder.writers, NewConsoleWriter().WithFormatter(builder.formatter))
	return builder
}

func (builder *LoggerBuilder) WithGrafanaLokiOutput(
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
	loki.WithFormatter(builder.formatter)
	builder.writers = append(builder.writers, loki)
	return builder
}

func (builder *LoggerBuilder) GetLogger() *Logger {
	return &Logger{
		level:   builder.level,
		writers: builder.writers,
	}
}
