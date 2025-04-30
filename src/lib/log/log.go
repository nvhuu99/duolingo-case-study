package log

import (
	"fmt"
	"strings"
	"time"
)

type Log struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      LogLevel  `json:"level"`
	LevelName  string    `json:"level_name"`
	URI        string    `json:"uri"`
	Message    string    `json:"message"`
	LogData    any       `json:"data"`
	LogErrors  any       `json:"errors"`
	LogContext any       `json:"context"`

	ready chan bool `json:"-"`
}

func NewLog(level LogLevel, message string) (*Log, <-chan bool) {
	ready := make(chan bool, 1)
	log := &Log{
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
	log.Timestamp = time.Now()

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

func (log *Log) GetStr(path string) (string, error) {
	data := map[string]any{
		"data":    log.LogData,
		"context": log.LogContext,
	}
	val, err := travel(path, data)
	if err == nil {
		if asStr, ok := val.(string); ok {
			return asStr, nil
		}
	}
	return "", err
}

func (log *Log) GetInt(path string) (int64, error) {
	data := map[string]any{
		"data":    log.LogData,
		"context": log.LogContext,
	}
	val, err := travel(path, data)
	if err == nil {
		if asNum, ok := val.(float64); ok {
			return int64(asNum), nil
		}
	}
	return 0, err
}

func (log *Log) GetFloat(path string) (float64, error) {
	data := map[string]any{
		"data":    log.LogData,
		"context": log.LogContext,
	}
	val, err := travel(path, data)
	if err == nil {
		if asNum, ok := val.(float64); ok {
			return asNum, nil
		}
	}
	return 0, err
}

func (log *Log) GetBool(path string) (bool, error) {
	data := map[string]any{
		"data":    log.LogData,
		"context": log.LogContext,
	}
	val, err := travel(path, data)
	if err == nil {
		if found, ok := val.(bool); ok {
			return found, nil
		}
	}
	return false, err
}

func (log *Log) GetRaw(path string) (any, error) {
	data := map[string]any{
		"data":    log.LogData,
		"context": log.LogContext,
	}
	found, err := travel(path, data)
	if err == nil {
		return found, nil
	}
	return nil, err
}

func travel(path string, data map[string]any) (any, error) {
	parts := strings.Split(path, ".")
	iterator := data
	for i := 0; i < len(parts); i++ {
		if _, exists := iterator[parts[i]]; !exists {
			break
		}
		if i == len(parts)-1 {
			return iterator[parts[i]], nil
		}
		next, ok := iterator[parts[i]].(map[string]any)
		if !ok {
			break
		}
		iterator = next
	}

	return nil, fmt.Errorf("\"%v\" not exists", path)
}
