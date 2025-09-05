package log

import (
	"time"
)

type Log struct {
	Timestamp    time.Time `json:"timestamp"`
	Level        LogLevel  `json:"level"`
	LevelName    string    `json:"level_name"`
	Message      string    `json:"message"`
	LogNamespace string    `json:"namespace"`
	LogData      any       `json:"data"`
	LogError     any       `json:"error"`
	LogContext   any       `json:"context"`
}

func NewLog(level LogLevel, message string) *Log {
	log := &Log{
		Level:     level,
		LevelName: logLevelAsString[level],
		Message:   message,
	}
	return log
}

func (log *Log) Namespace(namespace string) *Log {
	log.LogNamespace = namespace
	return log
}

func (log *Log) Context(ctx any) *Log {
	log.LogContext = ctx
	return log
}

func (log *Log) Data(data any) *Log {
	log.LogData = data
	return log
}

func (log *Log) Err(errs any) *Log {
	if asErr, ok := errs.(error); ok {
		log.LogError = asErr.Error()
	}
	return log
}

func (log *Log) Detail(detail map[string]any) *Log {
	if logCtx, has := detail["context"]; has {
		log.Context(logCtx)
	}
	if logData, has := detail["data"]; has {
		log.Data(logData)
	}
	if err, has := detail["err"]; has {
		log.Err(err)
	}
	return log
}
