package log

import (
	"context"
	lw "duolingo/lib/log/writer"
)

type Logger struct {
	formatter Formatter
	writers   []lw.LogWriter
	ctx       context.Context

	FilePrefix string
	Namespace string
}

func (logger *Logger) Info(message string) *Log {
	log, ready := NewLog(LevelInfo, message)
	log.Namespace = logger.Namespace
	logger.writeWhenReady(ready, log)

	return log
}

func (logger *Logger) Error(message string, errs any) *Log {
	log, ready := NewLog(LevelError, message)
	log.Namespace = logger.Namespace
	log.Errors(errs)
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

