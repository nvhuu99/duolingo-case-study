package log

import (
	"context"
	format "duolingo/lib/log/formatter"
	local "duolingo/lib/log/writer/driver/local"
	lw "duolingo/lib/log/writer"
	"time"
)

type Logger struct {
	formatter Formatter
	writers   []lw.LogWriter
	ctx       context.Context

	FilePrefix string
}

func NewLogger(ctx context.Context, filePrefix string) *Logger {
	logger := new(Logger)
	logger.ctx = ctx
	logger.FilePrefix = filePrefix
	logger.formatter = new(format.JsonFormatter)
	return logger
}

func (logger *Logger) UseJsonFormat() *Logger {
	logger.formatter = new(format.JsonFormatter)
	return logger
}

func (logger *Logger) AddLocalWriter(path string, bufferMb int, bufferCount int, rotation time.Duration) *Logger {
	writer := local.NewLocalWriter(logger.ctx, path, bufferMb, bufferCount, rotation)
	logger.writers = append(logger.writers, writer)
	return logger
}

func (logger *Logger) Info(message string) *Log {
	log, ready := NewLog(LevelInfo, message)

	logger.writeWhenReady(ready, log)

	return log
}

func (logger *Logger) Error(message string, errs... any) *Log {
	log, ready := NewLog(LevelError, message)

	log.Errors(errs...)

	logger.writeWhenReady(ready, log)

	return log
}

func (logger *Logger) writeWhenReady(ready <-chan bool, log *Log) {
	go func() {
		<-ready
		for _, writer := range logger.writers {
			writer.Write(&lw.Writable{
				Prefix: logger.FilePrefix,
				Extension: levelFileExtensions[log.Level],
				Content: logger.formatter.Format(log),
			})
		}
	}()
}

