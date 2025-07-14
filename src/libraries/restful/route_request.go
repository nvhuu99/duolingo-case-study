package restful

import (
	"duolingo/libraries/restful/router"
	"encoding/json"
	"io"

	"github.com/tidwall/gjson"
)

type routeRequest struct {
	BasePipeline
	*router.Router
}

func (r *routeRequest) Handle(req *Request, res *Response) {
	routeResult := r.route(req, res)
	r.setRequestHandler(req, res, routeResult)
	r.setRequestProperties(req, routeResult)
	r.Next(req, res)
}

func (r *routeRequest) route(req *Request, res *Response) *router.RouteResult {
	found, routeResult := r.Match(req.URL().Path)
	if !found {
		res.NotFound("uri not found: " + req.URL().Path)
		panic("uri not found: " + req.URL().Path)
	}
	return routeResult
}

func (r *routeRequest) setRequestProperties(
	req *Request,
	routeResult *router.RouteResult,
) {
	r.setRequestPathArgs(req, routeResult.GetPathArgs())
	r.setRequestQueries(req)
	r.setRequestBody(req)
}

func (r *routeRequest) setRequestPathArgs(req *Request, args map[string]string) {
	if js, jsErr := json.Marshal(args); jsErr == nil {
		req.pathArgs = gjson.ParseBytes(js)
	}
}

func (r *routeRequest) setRequestQueries(req *Request) {
	queriesMap := map[string][]string(req.base.URL.Query())
	reducedQueriesMap := make(map[string]any)
	for q := range queriesMap {
		if len(queriesMap[q]) == 1 {
			reducedQueriesMap[q] = queriesMap[q][0]
		}
		if len(queriesMap[q]) > 1 {
			reducedQueriesMap[q] = queriesMap[q]
		}
	}
	if js, jsErr := json.Marshal(reducedQueriesMap); jsErr == nil {
		req.queries = gjson.ParseBytes(js)
	}
}

func (r *routeRequest) setRequestBody(req *Request) {
	rawBody, readErr := io.ReadAll(req.base.Body)
	req.base.Body.Close()
	if readErr == nil {
		req.rawBody = rawBody
		req.inputs = gjson.ParseBytes(rawBody)
	}
}

func (r *routeRequest) setRequestHandler(
	req *Request,
	res *Response,
	routeResult *router.RouteResult,
) {
	reqHandler, ok := routeResult.GetHandler().(func(*Request, *Response))
	if !ok {
		res.ServerErr("invalid request handler")
		panic("invalid request handler")
	}
	req.handler = reqHandler
}
