package log_detail

import (
	cnst "duolingo/common/constant"
	"duolingo/lib/log"
	noti "duolingo/lib/notification"
	"duolingo/model"
	lc "duolingo/model/log_context"
)

type SendNotification struct {
	log.Log

	LogData struct {
		PushMessage *model.PushNotiMessage `json:"push_message"`
		PushResult  *noti.Result           `json:"push_result"`
	} `json:"data"`

	ContextAttr struct {
		RequestId string             `json:"request_id"`
		MessageId string             `json:"message_id"`
		Service   *lc.ServiceContext `json:"service"`
	} `json:"context"`
}

func SendNotificationDetail(message *model.PushNotiMessage, result *noti.Result) map[string]any {
	serviceContext := &lc.ServiceContext{
		Type:            cnst.ServiceTypes[cnst.SV_PUSH_SENDER],
		Name:            cnst.SV_PUSH_SENDER,
		Operation:       cnst.SEND_PUSH_NOTI,
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
			"request_id": mesgId,
			"message_id": reqId,
			"service":    serviceContext,
		},
		"data": map[string]any{
			"push_message": message,
			"push_result":  result,
		},
	}
}
