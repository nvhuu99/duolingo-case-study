package log

import "time"

type Log struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      LogLevel  `json:"level"`
	LevelName  string    `json:"level_name"`
	Namespace  string    `json:"namespace"`
	Message    string    `json:"message"`
	LogData    any       `json:"data"`
	LogErrors  any       `json:"errors"`
	LogContext any       `json:"context"`

	ready chan bool
}

func NewLog(level LogLevel, message string) (*Log, <-chan bool) {
	ready := make(chan bool, 1)
	log := &Log{
		Timestamp: time.Now(),
		Level:     level,
		LevelName: LogLevelAsString[level],
		Message:   message,
		ready:     ready,
	}
	return log, ready
}

func (log *Log) Context(attrs any) *Log {
	log.LogContext = attrs
	return log
}

func (log *Log) Data(attrs any) *Log {
	log.LogData = attrs
	return log
}

func (log *Log) Errors(errs any) *Log {
	if asErr, ok := errs.(error); ok {
		log.LogErrors = asErr.Error()
	}
	log.LogErrors = errs
	return log
}

func (log *Log) Write() {
	log.ready <- true
	close(log.ready)
}

func (log *Log) Detail(detail map[string]any) *Log {
	if logCtx, has := detail["context"]; has {
		log.Context(logCtx)
	}
	if logData, has := detail["data"]; has {
		log.Data(logData)
	}
	if errs, has := detail["errors"]; has {
		log.Errors(errs)
	}
	return log
}
