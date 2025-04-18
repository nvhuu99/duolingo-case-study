package detail

import (
	ed "duolingo/event/event_data"
	"duolingo/lib/log"
	lc "duolingo/model/log/context"
	ld "duolingo/model/log/data"
	"time"
)

type InputMessageRequest struct {
	log.Log

	LogContext struct {
		Trace *lc.TraceSpan `json:"trace"`
	} `json:"context"`

	LogData struct {
		Request *ld.HttpRequest `json:"request"`
	} `json:"data"`
}

func InpMsgRequestDetail(eventData *ed.InputMessageRequest, trace *lc.TraceSpan) map[string]any {
	request := eventData.Request
	response := eventData.Response
	return map[string]any{
		"context": map[string]any{
			"trace": trace,
		},
		"data": map[string]any{
			"request": &ld.HttpRequest{
				RequestId:        request.Id(),
				Timestamp:        request.Timestamp.UTC().Format(time.RFC3339),
				Method:           request.Method(),
				URL:              request.URL().Path,
				StatusCode:       response.Status,
				ClientAddr:       request.Instance().RemoteAddr,
				UserAgent:        request.Instance().UserAgent(),
				Referer:          request.Instance().Referer(),
				Query:            request.Query("*").Raw(),
				Inputs:           request.Input("*").Raw(),
				Headers:          request.Header(),
				ResponseTimeMs:   response.ResponseTimeMs,
				ResponseBodySize: response.ResponseBodySize,
				ResponseBodyData: response.Data,
			},
		},
	}
}
