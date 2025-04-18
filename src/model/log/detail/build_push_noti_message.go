package detail

import (
	ed "duolingo/event/event_data"
	"duolingo/lib/log"
	wd "duolingo/lib/work_distributor"
	"duolingo/model"
	lc "duolingo/model/log/context"
)

type BuildNotification struct {
	log.Log

	LogData *struct {
		Message      *model.InputMessage `json:"message"`
		Assignments  *wd.Assignment      `json:"assignments"`
		Workload     *wd.Workload        `json:"workload"`
		DeviceTokens []string            `json:"device_tokens"`
	} `json:"data"`

	LogContext struct {
		Trace *lc.TraceSpan `json:"trace"`
	} `json:"context"`
}

func BuildNotificationDetail(eventData *ed.BuildPushNotiMessage, trace *lc.TraceSpan) map[string]any {
	data := make(map[string]any)
	data["message"] = eventData.PushNoti.InputMessage
	data["assignents"] = eventData.Assignments
	data["workload"] = eventData.Workload
	if !eventData.Success {
		data["device_tokens"] = eventData.PushNoti.DeviceTokens
	}
	return map[string]any{
		"context": map[string]any{
			"trace": trace,
		},
		"data": data,
	}
}
