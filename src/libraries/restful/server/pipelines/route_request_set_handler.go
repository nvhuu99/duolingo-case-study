package pipelines

import (
	"duolingo/libraries/restful"
	"duolingo/libraries/restful/router"
)

type RouteRequestSetHandler struct {
	restful.BasePipeline
	*router.Router
}

func (r *RouteRequestSetHandler) Handle(req *restful.Request, res *restful.Response) {
	if routeResult := r.route(req, res); routeResult != nil {
		requestBuilder := restful.NewRequestBuilder(req)
		requestBuilder.SetPathArgs(routeResult.GetPathArgs())
		requestBuilder.SetHandler(routeResult.GetHandler())
		requestBuilder.Build()
	}
	r.Next(req, res)
}

func (r *RouteRequestSetHandler) route(
	req *restful.Request,
	res *restful.Response,
) *router.RouteResult {
	found, routeResult := r.Match(req.Method() + "/" + req.URL().Path)
	if !found {
		res.NotFound("uri not found: " + req.URL().Path)
		return nil
	}
	return routeResult
}
