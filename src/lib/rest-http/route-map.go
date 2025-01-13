package resthttp

type RouteMap struct {
	childs  map[string]*RouteMap
	handler *Handler
	pattern string
}
