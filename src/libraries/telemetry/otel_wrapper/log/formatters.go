package log

import (
	"encoding/json"
	"fmt"
)

type LogFormatter interface {
	Format(log *Log) (string, error)
}

/* Json Formatter */

type JsonFormatter struct {
}

func (formatter *JsonFormatter) Format(log *Log) (string, error) {
	b, err := json.Marshal(log)
	return string(b), err
}

/* Key-value Pair Formatter */

type KeyValuePairFormatter struct {
}

func (formatter *KeyValuePairFormatter) Format(log *Log) (string, error) {
	output := fmt.Sprintf(
		"%v level: %v - ns: %v - message: %v",
		log.Timestamp.Format("20060102150405"),
		LogLevelAsString(log.Level),
		log.LogNamespace,
		log.Message,
	)
	if log.LogError != nil {

	}
	return output, nil
}
