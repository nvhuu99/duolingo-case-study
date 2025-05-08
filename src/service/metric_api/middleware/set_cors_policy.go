package middleware

import (
	rest "duolingo/lib/rest_http"
)

type SetCORSPolicies struct {
	rest.BaseMiddleware
}

func (mw *SetCORSPolicies) Handle(request *rest.Request, response *rest.Response) {
	defer mw.Next(request, response)
	response.Header().Set("Access-Control-Allow-Origin", "*")
}
