package resthttp

type RouteMap struct {
	name    string
	childs  map[string]*RouteMap
	handler *Handler
	pattern string
}
