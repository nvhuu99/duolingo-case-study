package json

import (
	"encoding/json"
)

type JsonFormatter struct {
}

func (formatter *JsonFormatter) Format(log any) []byte {
	out, _ := json.Marshal(log)
	return out
}
