package model

import (
	"encoding/json"

	"github.com/google/uuid"
)

type PushNotiMessage struct {
	Id           string   `json:"id"`
	RequestId    string   `json:"request_id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	DeviceTokens []string `json:"device_tokens"`
}

func NewPushNotiMessage(requestId string, title string, content string, deviceTokens []string) *PushNotiMessage {
	return &PushNotiMessage{
		Id:           uuid.New().String(),
		RequestId:    requestId,
		Title:        title,
		Content:      content,
		DeviceTokens: deviceTokens,
	}
}

func (msg *PushNotiMessage) Serialize() string {
	s, _ := json.Marshal(msg)
	return string(s)
}
