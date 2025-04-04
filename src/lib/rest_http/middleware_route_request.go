package rest_http

type RouteRequest struct {
	BaseMiddleware
	server *Server
}

func (mw *RouteRequest) Handle(request *Request, response *Response) {
	defer mw.Next(request, response)

	if handler := mw.server.Router().Match(request); handler != nil {
		handler(request, response)
	} else {
		response.NotFound("")
		mw.Terminate()
	}
}
