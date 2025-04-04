package middleware

import (
	rest "duolingo/lib/rest_http"
)

type LogResponse struct {
	rest.BaseMiddleware
}

func (mw *LogResponse) Handle(request *rest.Request, response *rest.Response) {
	defer mw.Next(request, response)
}
