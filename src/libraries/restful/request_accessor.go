package restful

import "net/http"

type RequestAccessor struct {
	request *Request
}

func NewRequestAccessor(req *Request) *RequestAccessor {
	return &RequestAccessor{request: req}
}

func (accessor *RequestAccessor) GetBaseObject() *http.Request {
	return accessor.request.base
}

func (accessor *RequestAccessor) GetRequestHandler() func(*Request, *Response) {
	return accessor.request.handler
}
