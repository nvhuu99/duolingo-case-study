package log

import (
	jf "duolingo/lib/log/driver/formatter/json"
	lw "duolingo/lib/log/driver/writer/local"
	"time"

	"context"
)

type LoggerBuilder struct {
	logger *Logger
}

func NewLoggerBuilder(ctx context.Context) *LoggerBuilder {
	logger := new(Logger)
	logger.ctx = ctx
	logger.formatter = new(jf.JsonFormatter)
	return &LoggerBuilder{
		logger: logger,
	}
}

func (builder *LoggerBuilder) Get() *Logger {
	return builder.logger
}

func (builder *LoggerBuilder) SetLogLevel(level LogLevel) *LoggerBuilder {
	builder.logger.level = level
	return builder
}

func (builder *LoggerBuilder) UseNamespace(parts ...string) *LoggerBuilder {
	builder.logger.Namespace = Namespace(parts...)
	return builder
}

func (builder *LoggerBuilder) UseJsonFormat() *LoggerBuilder {
	builder.logger.formatter = new(jf.JsonFormatter)
	return builder
}

func (builder *LoggerBuilder) UseLocalWriter(path string) *LoggerBuilder {
	writer := lw.NewLocalWriter(builder.logger.ctx, path)
	builder.logger.writer = writer
	return builder
}

func (builder *LoggerBuilder) WithFilePrefix(prefix string) *LoggerBuilder {
	builder.logger.FilePrefix = prefix
	return builder
}

func (builder *LoggerBuilder) WithBuffering(sizeMb int, maxCount int) *LoggerBuilder {
	builder.logger.writer.WithBuffering(sizeMb, maxCount)
	return builder
}

func (builder *LoggerBuilder) WithRotation(interval time.Duration) *LoggerBuilder {
	builder.logger.writer.WithRotation(interval)
	return builder
}

func (builder *LoggerBuilder) WithFlushInterval(interval time.Duration) *LoggerBuilder {
	builder.logger.writer.WithFlushInterval(interval)
	return builder
}
