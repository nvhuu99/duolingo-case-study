package rest_http

type Handler interface {
	SetNext(Handler)
	Handle(*Request, *Response)
}
