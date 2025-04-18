package event_data

import "duolingo/model"

type RelayInputMessage struct {
	OptId               string
	PushNoti            *model.PushNotiMessage
	Success             bool
	Error               error
	RelayedCount        uint8
	MessageIgnoreReason string
}
