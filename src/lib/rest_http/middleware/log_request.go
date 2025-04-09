package middleware

import (
	log "duolingo/lib/log"
	rest "duolingo/lib/rest_http"
)

type LogRequest struct {
	rest.BaseMiddleware

	logger *log.Logger
}

func (mw *LogRequest) Handle(request *rest.Request, response *rest.Response) {
	defer mw.Next(request, response)
	mw.logger.Info("receive new message").Context(request).Write()
}
