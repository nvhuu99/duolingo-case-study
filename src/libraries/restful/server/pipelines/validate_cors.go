package pipelines

import (
	"duolingo/libraries/restful"
	"net/http"
	"slices"
)

type ValidateCORS struct {
	restful.BasePipeline
}

func (pipeline *ValidateCORS) Handle(req *restful.Request, res *restful.Response) {
	pipeline.checkMethod(req, res)
	pipeline.checkContentType(req, res)
	if !res.Sent() {
		pipeline.Next(req, res)
	}
}

func (pipeline *ValidateCORS) checkMethod(req *restful.Request, res *restful.Response) {
	allowMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	if !slices.Contains(allowMethods, req.Method()) {
		res.Send(http.StatusMethodNotAllowed, false, "Method is not allowed", nil, nil)
	}
}

func (pipeline *ValidateCORS) checkContentType(req *restful.Request, res *restful.Response) {
	if req.Header().Get("Content-Type") != "application/json" {
		res.Send(http.StatusUnsupportedMediaType, false, "Only 'application/json' content type is supported", nil, nil)
	}
}
