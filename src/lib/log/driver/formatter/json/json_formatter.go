package json

import (
	"encoding/json"
)

type JsonFormatter struct {
}

func (formatter *JsonFormatter) Format(log any) ([]byte, error) {
	return json.Marshal(log)
}
