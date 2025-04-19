package log

import (
	"context"
	fm "duolingo/lib/log/formatter"
	lw "duolingo/lib/log/writer"
)

type Logger struct {
	formatter fm.Formatter
	writer    lw.LogWriter
	ctx       context.Context
	level     LogLevel

	FilePrefix string
	Namespace  string
}

func (logger *Logger) Info(message string) *Log {
	log, ready := NewLog(LevelInfo, message)
	log.Namespace = logger.Namespace
	logger.writeWhenReady(ready, log)

	return log
}

func (logger *Logger) Debug(message string) *Log {
	log, ready := NewLog(LevelDebug, message)
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

		if logger.level&log.Level != 0 {
			formatted, _ := logger.formatter.Format(log)
			logger.writer.Write(&lw.Writable{
				Namespace: log.Namespace,
				Prefix:    logger.FilePrefix,
				Extension: LogLevelAsString[log.Level],
				Content: formatted,
			})
		}
	}()
}
