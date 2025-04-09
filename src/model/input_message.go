package model

import (
	"encoding/json"

	"github.com/google/uuid"
)

type InputMessage struct {
	Id        string `json:"id"`
	RequestId string `json:"request_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	IsRelayed bool   `json:"isRelayed"` // Is this a relayed one or the original message
	Campaign  string `json:"campaign"`
}

func NewInputMessage(requestId string, campaign string, title string, content string, relayed bool) *InputMessage {
	return &InputMessage{
		Id:        uuid.New().String(),
		RequestId: requestId,
		Campaign:  campaign,
		Title:     title,
		Content:   content,
		IsRelayed: relayed,
	}
}

func (msg *InputMessage) Serialize() string {
	s, _ := json.Marshal(msg)
	return string(s)
}
