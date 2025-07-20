package pipelines

import (
	"duolingo/libraries/restful"
	"log"
)

type ExecuteRequestHandler struct {
	restful.BasePipeline
}

func (r *ExecuteRequestHandler) Handle(req *restful.Request, res *restful.Response) {
	defer r.panicHandler(res)
	accessor := restful.NewRequestAccessor(req)
	if handler := accessor.GetRequestHandler(); handler != nil {
		handler(req, res)
		log.Println("request handled", req.Method(), req.URL().Path)
	}
	r.Next(req, res)
}

func (r *ExecuteRequestHandler) panicHandler(response *restful.Response) {
	if r := recover(); r != nil {
		log.Println("failed to handle request", r)
		response.ServerErr("failed to handle request")
	}
}
