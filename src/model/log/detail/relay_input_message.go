package detail

import (
	ed "duolingo/event/event_data"
	"duolingo/lib/log"
	"duolingo/model"
	lc "duolingo/model/log/context"
)

type RelayInputMessage struct {
	log.Log

	LogData struct {
		MessageIgnored       bool                `json:"message_ignored"`
		MessageIgnoredReason string              `json:"message_ignored_reason"`
		RelayedMessage       *model.InputMessage `json:"relayed_message"`
		RelayedCount         int                 `json:"relayed_count"`
	} `json:"data"`

	LogContext struct {
		Trace *lc.TraceSpan `json:"trace"`
	} `json:"context"`
}

func RelayInpMsgDetail(result *ed.RelayInputMessage, trace *lc.TraceSpan) map[string]any {
	data := make(map[string]any)
	if result.Success {
		if result.MessageIgnoreReason != "" {
			data["message_ignored"] = true
			data["message_ignored_reason"] = result.MessageIgnoreReason
		} else {
			data["relayed_message"] = result.PushNoti.InputMessage
			data["relayed_count"] = result.RelayedCount
		}
	}
	return map[string]any{
		"context": map[string]any{
			"trace": trace,
		},
		"data": data,
	}
}
