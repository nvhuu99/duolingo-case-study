package detail

import (
	ed "duolingo/event/event_data"
	"duolingo/lib/log"
	"duolingo/model"
	lc "duolingo/model/log/context"
)

type SendPushNotification struct {
	log.Log

	LogData struct {
		Message       *model.InputMessage `json:"message"`
		Success       bool                `json:"success"`
		SuccessCount  int                 `json:"success_count"`
		FailureCount  int                 `json:"failure_count"`
		FailureTokens []string            `json:"failure_tokens"`
	} `json:"data"`

	LogContext struct {
		Trace *lc.TraceSpan `json:"trace"`
	} `json:"context"`
}

func SendPushNotiDetail(eventData *ed.SendPushNotification, trace *lc.TraceSpan) map[string]any {
	return map[string]any{
		"context": map[string]any{
			"trace": trace,
		},
		"data": map[string]any{
			"message":        eventData.PushNoti.InputMessage,
			"success":        eventData.Result.Success,
			"success_count":  eventData.Result.SuccessCount,
			"failure_count":  eventData.Result.FailureCount,
			"failure_tokens": eventData.Result.FailureTokens,
		},
	}
}
