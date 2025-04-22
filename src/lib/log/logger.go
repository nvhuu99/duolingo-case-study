package log

import (
	"context"
	fm "duolingo/lib/log/formatter"
	lw "duolingo/lib/log/writer"
)

type Logger struct {
	formatter fm.Formatter
	writer    *lw.LogWriter
	ctx       context.Context
	level     LogLevel
	uri       string
}

func (logger *Logger) Info(message string) *Log {
	log, ready := NewLog(LevelInfo, message)
	log.URI = logger.uri
	logger.writeWhenReady(ready, log)

	return log
}

func (logger *Logger) Debug(message string) *Log {
	log, ready := NewLog(LevelDebug, message)
	log.URI = logger.uri
	logger.writeWhenReady(ready, log)

	return log
}

func (logger *Logger) Error(message string, errs any) *Log {
	log, ready := NewLog(LevelError, message)
	log.URI = logger.uri
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
				URI:     log.URI + "." + LogLevelAsString[log.Level],
				Content: formatted,
			})
		}
	}()
}
