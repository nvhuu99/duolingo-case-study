package log_detail

import (
	cnst "duolingo/constant"
	"duolingo/lib/log"
	wd "duolingo/lib/work_distributor"
	"duolingo/model"
	lc "duolingo/model/log_context"
)

type BuildNotification struct {
	log.Log

	LogData *struct {
		PushMessage *model.InputMessage `json:"push_message"`
		Assignment  *wd.Assignment      `json:"assignment"`
		Workload    *wd.Workload        `json:"workload"`
	} `json:"data"`

	ContextAttr *struct {
		RequestId string             `json:"request_id"`
		MessageId string             `json:"message_id"`
		Service   *lc.ServiceContext `json:"service"`
	} `json:"context"`
}

func BuildNotificationDetail(message *model.InputMessage, workload *wd.Workload, assignment *wd.Assignment) map[string]any {
	serviceContext := &lc.ServiceContext{
		Type:            cnst.ServiceTypes[cnst.SV_NOTI_BUILDER],
		Name:            cnst.SV_NOTI_BUILDER,
		Operation:       cnst.BUILD_NOTI_MSG,
		InstanceId:      "",
		InstanceAddress: "",
	}

	mesgId := ""
	reqId := ""
	if message != nil {
		mesgId = message.Id
		reqId = message.RequestId
	}

	return map[string]any{
		"context": map[string]any{
			"request_id": reqId,
			"message_id": mesgId,
			"service":    serviceContext,
		},
		"data": map[string]any{
			"push_message": message,
			"assignment":   assignment,
			"workload":     workload,
		},
	}
}
