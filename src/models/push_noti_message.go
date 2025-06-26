package models

import (
	"encoding/json"
	"errors"
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

func (m *PushNotiMessage) Validate() error {
	if m.MessageInput == nil || m.MessageInput.Title == "" || m.MessageInput.Body == "" {
		return errors.New("push notification message is empty")
	}
	if len(m.TargetDevices) == 0 {
		return errors.New("push notification target devices is empty")
	}
	return nil
}

func (m *PushNotiMessage) GetTargetTokens(platforms []string) []string {
	tokens := []string{}
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
