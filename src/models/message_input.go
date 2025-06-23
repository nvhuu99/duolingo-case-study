package models

import "encoding/json"

type MessageInput struct {
	Id       string `json:"id"`
	Campaign string `json:"campaign"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

func NewMessageInput(campaign string, title string, body string) *MessageInput {
	return &MessageInput{
		Campaign: campaign,
		Title:    title,
		Body:     body,
	}
}

func (m *MessageInput) Encode() []byte {
	marshalled, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return marshalled
}

func MessageInputDecode(data []byte) *MessageInput {
	input := new(MessageInput)
	err := json.Unmarshal(data, input)
	if err != nil {
		panic(err)
	}
	return input
}
