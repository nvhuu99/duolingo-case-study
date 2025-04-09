package rest_http

type BaseMiddleware struct {
	next       Handler
	terminated bool
}

func (mw *BaseMiddleware) SetNext(handler Handler) {
	mw.next = handler
}

func (mw *BaseMiddleware) Terminate() {
	mw.terminated = true
}

func (mw *BaseMiddleware) Next(request *Request, response *Response) {
	if mw.next != nil && !mw.terminated {
		mw.next.Handle(request, response)
	}
}
