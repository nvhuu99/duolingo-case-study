package model

import (
	"encoding/json"
	"time"
)

type InputMessage struct {
	MessageId string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Campaign  string `json:"campaign"`
	CreatedAt time.Time `json:"created_at"`
}

func (msg *InputMessage) Serialize() string {
	s, _ := json.Marshal(msg)
	return string(s)
}
