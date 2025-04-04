package log

import "time"

type Log struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       LogLevel  `json:"level"`
	Namespace   string    `json:"namespace"`
	Message     string    `json:"message"`
	LogData     any       `json:"data"`
	LogErrors   any       `json:"errors"`
	GroupAttrs  any       `json:"group"`
	ContextAttr any       `json:"context"`

	ready chan bool
}

func NewLog(level LogLevel, message string) (*Log, <-chan bool) {
	ready := make(chan bool, 1)
	log := &Log{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		ready:     ready,
	}
	return log, ready
}

func (log *Log) Group(namespace string, attrs any) *Log {
	log.Namespace = namespace
	log.GroupAttrs = attrs
	return log
}

func (log *Log) Context(attrs any) *Log {
	log.ContextAttr = attrs
	return log
}

func (log *Log) Data(attrs any) *Log {
	log.LogData = attrs
	return log
}

func (log *Log) Errors(errs ...any) *Log {
	log.LogErrors = errs
	return log
}

func (log *Log) Write() {
	log.ready <- true
	close(log.ready)
}
