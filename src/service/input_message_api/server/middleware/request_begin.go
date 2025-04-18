package middleware

// import (
// 	rest "duolingo/lib/rest_http"
// 	sv "duolingo/lib/service_container"
// 	ep "duolingo/lib/event"
// 	ed "duolingo/event/event_data"
// 	eh "duolingo/event/event_handler"
// )

// type RequestBegin struct {
// 	rest.BaseMiddleware
// }

// func (mw *RequestBegin) Handle(request *rest.Request, response *rest.Response) {
// 	defer mw.Next(request, response)

// 	container := sv.GetContainer()

// 	container.BindSingleton("server.request", func() any { return request })

// 	container.BindSingleton("server.response", func() any { return response })

// }
