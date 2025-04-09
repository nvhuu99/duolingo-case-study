package middleware

import (
	log "duolingo/lib/log"
	rest "duolingo/lib/rest_http"
	ld "duolingo/model/log_detail"
)

type LogHandledRequest struct {
	rest.BaseMiddleware
	Logger *log.Logger
}

func (mw *LogHandledRequest) Handle(request *rest.Request, response *rest.Response) {
	defer mw.Next(request, response)

	logDetail := ld.InputMessageRequestDetail(request, response)

	if response.Errors != nil {
		mw.Logger.Error("request error", response.Errors).Detail(logDetail).Write()
	} else {
		mw.Logger.Info("request handled").Detail(logDetail).Write()
	}
}
