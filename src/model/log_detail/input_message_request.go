package log_detail

import (
	cnst "duolingo/constant"
	"duolingo/lib/log"
	rest "duolingo/lib/rest_http"
	"duolingo/model"
	lc "duolingo/model/log_context"
	"time"
)

type InputMessageRequest struct {
	log.Log

	LogData *model.InputMessage `json:"data"`

	ContextAttr struct {
		Request *lc.RequestContext `json:"request"`
		Service *lc.ServiceContext `json:"service"`
	} `json:"context"`
}

func InputMessageRequestDetail(request *rest.Request, response *rest.Response) map[string]any {
	requestContext := &lc.RequestContext{
		RequestId:        request.Id(),
		Timestamp:        request.Timestamp.UTC().Format(time.RFC3339),
		Method:           request.Method(),
		URL:              request.URL().Path,
		StatusCode:       response.Status,
		ClientAddr:       request.Instance().RemoteAddr,
		UserAgent:        request.Instance().UserAgent(),
		Referer:          request.Instance().Referer(),
		ResponseTimeMs:   response.ResponseTimeMs,
		ResponseBodySize: response.ResponseBodySize,
		Query:            request.Query("*").Raw(),
		Inputs:           request.Input("*").Raw(),
		Headers:          request.Header(),
	}

	serviceContext := &lc.ServiceContext{
		Type:            cnst.ServiceTypes[cnst.SV_INP_MESG],
		Name:            cnst.SV_INP_MESG,
		Operation:       cnst.INP_MESG_REQUEST,
		InstanceId:      "",
		InstanceAddress: "",
	}

	return map[string]any{
		"context": map[string]any{
			"request": requestContext,
			"service": serviceContext,
		},
		"data": response.Data,
	}
}
