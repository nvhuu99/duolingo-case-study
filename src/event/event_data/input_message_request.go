package event_data

import (
	"duolingo/lib/rest_http"
	"duolingo/model"
)

type InputMessageRequest struct {
	OptId    string
	PushNoti *model.PushNotiMessage
	Request  *rest_http.Request
	Response *rest_http.Response
	Success  bool
	Error    error
}
