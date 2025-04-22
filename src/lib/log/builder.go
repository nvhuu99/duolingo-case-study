package log

import (
	jf "duolingo/lib/log/driver/formatter/json"
	lf "duolingo/lib/log/driver/writer/local_file"
	lw "duolingo/lib/log/writer"
	"time"

	"context"
)

type LoggerBuilder struct {
	logger *Logger
}

func NewLoggerBuilder(ctx context.Context) *LoggerBuilder {
	return &LoggerBuilder{
		logger: &Logger{
			ctx: ctx,
			formatter: new(jf.JsonFormatter),
			writer: lw.NewLogWriter(ctx),
		},
	}
}

func (builder *LoggerBuilder) Get() *Logger {
	return builder.logger
}

func (builder *LoggerBuilder) SetLogLevel(level LogLevel) *LoggerBuilder {
	builder.logger.level = level
	return builder
}

func (builder *LoggerBuilder) SetURI(uri string) *LoggerBuilder {
	builder.logger.uri = uri
	return builder
}

func (builder *LoggerBuilder) UseJsonFormat() *LoggerBuilder {
	builder.logger.formatter = new(jf.JsonFormatter)
	return builder
}

func (builder *LoggerBuilder) WithLocalFileOutput(dir string) *LoggerBuilder {
	builder.logger.writer.AddLogOutput(lf.NewLocalFileOutPut(dir))
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

func (builder *LoggerBuilder) WithFlushInterval(interval time.Duration, grace time.Duration) *LoggerBuilder {
	builder.logger.writer.WithFlushInterval(interval, grace)
	return builder
}
