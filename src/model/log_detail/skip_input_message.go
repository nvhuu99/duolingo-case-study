package log_detail

import (
	cnst "duolingo/constant"
	"duolingo/lib/log"
	"duolingo/model"
	lc "duolingo/model/log_context"
)

type SkipInputMessage struct {
	log.Log

	LogData struct {
		SkippedMessage *model.InputMessage `json:"skipped_message"`
		SkippedReason  string              `json:"skipped_reason"`
	} `json:"data"`

	ContextAttr struct {
		RequestId string             `json:"request_id"`
		MessageId string             `json:"message_id"`
		Service   *lc.ServiceContext `json:"service"`
	} `json:"context"`
}

func SkipInputMessageDetail(message *model.InputMessage, skipReason string) map[string]any {
	serviceContext := &lc.ServiceContext{
		Type:            cnst.ServiceTypes[cnst.SV_NOTI_BUILDER],
		Name:            cnst.SV_NOTI_BUILDER,
		Operation:       cnst.SKIP_INP_MESG,
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
			"skipped_message": message,
			"skipped_reason":  skipReason,
		},
	}
}
