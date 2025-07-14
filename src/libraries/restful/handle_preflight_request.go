package restful

import (
	"net/http"
)

type handleRequestPreflight struct {
	BasePipeline
}

func (pipeline *handleRequestPreflight) Handle(req *Request, res *Response) {
	if req.Method() == http.MethodOptions {
		res.SetHeader("Allow", "GET, POST, PUT, DELETE, OPTIONS")
		res.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		res.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
		res.SetHeader("Access-Control-Allow-Origin", "*")
		res.SetHeader("Access-Control-Max-Age", "86400")
		res.NoContent()
	} else {
		pipeline.Next(req, res)
	}
}
