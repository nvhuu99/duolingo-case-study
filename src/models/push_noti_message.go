package models

import (
	"encoding/json"
	"slices"
)

type PushNotiMessage struct {
	*MessageInput
	TargetDevices []*UserDevice `json:"target_devices"`
}

func NewPushNotiMessage(input *MessageInput, devices []*UserDevice) *PushNotiMessage {
	return &PushNotiMessage{
		MessageInput:  input,
		TargetDevices: devices,
	}
}

func (m *PushNotiMessage) GetTargetTokens(platforms []string) []string {
	tokens := make([]string, len(m.TargetDevices))
	for i := range m.TargetDevices {
		if slices.Contains(platforms, m.TargetDevices[i].Platform) {
			if m.TargetDevices[i].Token != "" {
				tokens = append(tokens, m.TargetDevices[i].Token)
			}
		}
	}
	return tokens
}

func (m *PushNotiMessage) Encode() []byte {
	marshalled, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return marshalled
}

func PushNotiMessageDecode(data []byte) *PushNotiMessage {
	input := new(PushNotiMessage)
	err := json.Unmarshal(data, input)
	if err != nil {
		panic(err)
	}
	return input
}
