package rest_http

type RouteMap struct {
	name    string
	childs  map[string]*RouteMap
	handler func (*Request, *Response)
	pattern string
}
