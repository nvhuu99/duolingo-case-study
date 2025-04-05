package log

import (
	format "duolingo/lib/log/formatter"
	local "duolingo/lib/log/writer/driver/local"
	"time"

	"context"
)

type LoggerBuilder struct {
	logger *Logger
}

func NewLoggerBuilder(ctx context.Context) *LoggerBuilder {
	logger := new(Logger)
	logger.ctx = ctx
	logger.formatter = new(format.JsonFormatter)
	return &LoggerBuilder{
		logger: logger,
	}
}

func (builder *LoggerBuilder) Get() *Logger {
	return builder.logger
}

func (builder *LoggerBuilder) UseNamespace(parts... string) *LoggerBuilder {
	builder.logger.Namespace = Namespace(parts...)
	return builder
}

func (builder *LoggerBuilder) UseJsonFormat() *LoggerBuilder {
	builder.logger.formatter = new(format.JsonFormatter)
	return builder
}

func (builder *LoggerBuilder) AddLocalWriter(path string) *LoggerBuilder {
	writer := local.NewLocalWriter(builder.logger.ctx, path)
	builder.logger.writers = append(builder.logger.writers, writer)
	return builder
}

func (builder *LoggerBuilder) WithFilePrefix(prefix string) *LoggerBuilder {
	builder.logger.FilePrefix = prefix
	return builder
}

func (builder *LoggerBuilder) WithBuffering(sizeMb int, maxCount int) *LoggerBuilder {
	for _, writer := range builder.logger.writers {
		writer.WithBuffering(sizeMb, maxCount)
	}
	return builder
}

func (builder *LoggerBuilder) WithRotation(interval time.Duration) *LoggerBuilder {
	for _, writer := range builder.logger.writers {
		writer.WithRotation(interval)
	}
	return builder
}

func (builder *LoggerBuilder) WithFlushInterval(interval time.Duration) *LoggerBuilder {
	for _, writer := range builder.logger.writers {
		writer.WithFlushInterval(interval)
	}
	return builder
}

