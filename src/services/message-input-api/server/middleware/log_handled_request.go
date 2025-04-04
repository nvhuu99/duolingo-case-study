package middleware

import (
	log "duolingo/lib/log"
	rest "duolingo/lib/rest_http"
	"time"
)

type LogHandledRequest struct {
	rest.BaseMiddleware
	Logger *log.Logger
}

func (mw *LogHandledRequest) Handle(request *rest.Request, response *rest.Response) {
	defer mw.Next(request, response)

	handledRequest := requestInfo(request, response)
	
	if response.Errors != nil {
		mw.Logger.Error("request error", response.Errors).
			Context(handledRequest).
			Group(log.Namespace("services", "message_input_api"), nil).
			Write()
	} else {
		mw.Logger.Info("request handled").
			Context(handledRequest).
			Group(log.Namespace("services", "message_input_api"), nil).
			Write()
	}
}

func requestInfo(request *rest.Request, response *rest.Response) any {
	return &struct {
		RequestId        string `json:"request_id"`
		Timestamp        string `json:"timestamp"`
		Method           string `json:"method"`
		URL              string `json:"url"`
		StatusCode       int    `json:"status_code"`
		ClientAddr       string `json:"client_address"`
		UserAgent        string `json:"user_agent"`
		Referer          string `json:"referer"`
		ResponseTimeMs   int    `json:"response_time_ms"`
		ResponseBodySize int    `json:"response_body_size"`
		Query            any    `json:"query"`
		Inputs           any    `json:"inputs"`
		Headers          any    `json:"headers"`
	}{
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
}
