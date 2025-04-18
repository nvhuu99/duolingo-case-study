package middleware

// import (
// 	rest "duolingo/lib/rest_http"
// 	sv "duolingo/lib/service_container"
// 	ep "duolingo/lib/event"
// 	se "duolingo/service/input_message_api/event"
// )

// type RequestEnd struct {
// 	rest.BaseMiddleware
// }

// func (mw *RequestEnd) Handle(request *rest.Request, response *rest.Response) {
// 	defer mw.Next(request, response)

// 	container := sv.GetContainer()

// 	event := container.Resolve("event.publisher").(*ep.EventPublisher)

// 	container.BindSingleton("server.request", func() any { return request })

// 	container.BindSingleton("server.response", func() any { return response })

// 	event.Notify(se.INP_MSG_REQUEST_END, nil)
// }
