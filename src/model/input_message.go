package model

import (
	"encoding/json"
)

type InputMessage struct {
	MessageId string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"body"`
	Campaign  string `json:"campaign"`
}

func (msg *InputMessage) Serialize() string {
	s, _ := json.Marshal(msg)
	return string(s)
}
