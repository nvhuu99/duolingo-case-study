package log_detail

import (
	cnst "duolingo/constant"
	"duolingo/lib/log"
	"duolingo/model"
	lc "duolingo/model/log_context"
)

type RelayInputMessage struct {
	log.Log

	LogData struct {
		RelayedMessage *model.InputMessage `json:"relayed_message"`
		RelayedTotal   int                 `json:"relayed_total"`
	} `json:"data"`

	ContextAttr struct {
		RequestId string             `json:"request_id"`
		MessageId string             `json:"message_id"`
		Service   *lc.ServiceContext `json:"service"`
	} `json:"context"`
}

func RelayInputMessageDetail(message *model.InputMessage, relayTotal int) map[string]any {
	serviceContext := &lc.ServiceContext{
		Type:            cnst.ServiceTypes[cnst.SV_NOTI_BUILDER],
		Name:            cnst.SV_NOTI_BUILDER,
		Operation:       cnst.RELAY_INP_MESG,
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
			"relayed_message": message,
			"relayed_total":   relayTotal,
		},
	}
}
