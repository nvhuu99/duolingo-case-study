package log

import (
	"time"
)

type Logger struct {
	writers []LogWriter
	level   LogLevel
}

func (logger *Logger) Info(message string) *Log {
	log := NewLog(LevelInfo, message)
	return log
}

func (logger *Logger) Debug(message string) *Log {
	log := NewLog(LevelInfo, message)
	return log
}

func (logger *Logger) Error(message string, err any) *Log {
	log := NewLog(LevelInfo, message)
	log.Err(err)
	return log
}

func (logger *Logger) UnlessError(
	err error,
	messageIfErr string,
	level LogLevel,
	message string,
) *Log {
	log := NewLog(level, message)
	if err != nil {
		log.Level = LevelError
		log.Message = messageIfErr
		log.Err(err)
	} else {
		log.Level = level
		log.Message = message
	}
	return log
}

func (logger *Logger) Write(log *Log) {
	if logger.level >= log.Level {

		if log.LogError != nil {
			log.Level = LevelError
		}
		log.Timestamp = time.Now()

		for _, writer := range logger.writers {
			writer.Write(log)
		}
	}
}
