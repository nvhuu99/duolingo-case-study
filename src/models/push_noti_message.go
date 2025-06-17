package models

import "encoding/json"

type PushNotiMessage struct {
	*MessageInput
	DeviceTokens []string `json:"tokens"`
}

func NewPushNotiMessage(input *MessageInput, tokens []string) *PushNotiMessage {
	return &PushNotiMessage{
		MessageInput: input,
		DeviceTokens: tokens,
	}
}

func (m *PushNotiMessage) Encode() []byte {
	marshalled, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return marshalled
}
