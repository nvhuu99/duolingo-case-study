package restful

import (
	"errors"
	"net/http"
	"slices"
)

type validateCORS struct {
	BasePipeline
}

func (pipeline *validateCORS) Handle(req *Request, res *Response) {
	pipeline.checkMethod(req, res)
	pipeline.checkHeaders(req, res)
	pipeline.checkContentType(req, res)
	pipeline.Next(req, res)
}

func (pipeline *validateCORS) checkMethod(req *Request, res *Response) {
	allowMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	if !slices.Contains(allowMethods, req.Method()) {
		res.send(http.StatusMethodNotAllowed, false, "Method is not allowed", nil, nil)
		panic(errors.New("Method is not allowed"))
	}
}

func (pipeline *validateCORS) checkHeaders(req *Request, res *Response) {
	allowHeaders := []string{"Content-Type", "Authorization"}
	for header := range req.base.Header {
		if !slices.Contains(allowHeaders, header) {
			res.send(http.StatusForbidden, false, "Header is not allowed", nil, nil)
			panic(errors.New("Header is not allowed"))
		}
	}
}

func (pipeline *validateCORS) checkContentType(req *Request, res *Response) {
	if req.base.Header.Get("Content-Type") != "application/json" {
		res.send(http.StatusUnsupportedMediaType, false, "Only 'application/json' content type is supported", nil, nil)
		panic("Only 'application/json' content type is supported")
	}
}
