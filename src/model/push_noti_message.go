package model

import (
	"duolingo/model/log/context"
	"encoding/json"
)

type RelayFlag string

const (
	ShouldRelay RelayFlag = "should_relay"
	HasRelayed  RelayFlag = "has_relayed"
)

type PushNotiMessage struct {
	RelayFlag    RelayFlag          `json:"relay_flg"`
	InputMessage *InputMessage      `json:"msg"`
	Trace        *context.TraceSpan `json:"trace"`
	DeviceTokens []string           `json:"tkns"`
}

func (msg *PushNotiMessage) Serialize() string {
	s, _ := json.Marshal(msg)
	return string(s)
}
