package log

import (
	"encoding/json"
)

type LogFormatter interface {
	Format(log any) (string, error)
}

/* JsonFormatter */

type JsonFormatter struct {
}

func (formatter *JsonFormatter) Format(log any) (string, error) {
	b, err := json.Marshal(log)
	return string(b), err
}
