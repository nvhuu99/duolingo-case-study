package router

type RouteResult struct {
	routeNode   *routeNode
	requestPath string
	pathArgs    map[string]string
}

func newRouteResult(requestPath string, routeNode *routeNode) *RouteResult {
	return &RouteResult{
		routeNode:   routeNode,
		requestPath: requestPath,
		pathArgs:    extractPathArgs(routeNode.pattern, requestPath),
	}
}

func (r *RouteResult) GetHandler() any {
	return r.routeNode.handler
}

func (r *RouteResult) GetPathArgs() map[string]string {
	return r.pathArgs
}
