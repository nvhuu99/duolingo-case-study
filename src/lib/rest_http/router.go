package rest_http

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Router struct {
	routeMap *RouteMap
}

func NewRouter() *Router {
	router := Router{
		routeMap: &RouteMap{
			childs: make(map[string]*RouteMap),
		},
	}

	return &router
}

func (router *Router) Match(request *Request) func (*Request, *Response) {
	found, err := router.matchRoute(request.Method(), request.URL().Path)
	if err != nil {
		return nil
	}

	pathValues := parsePath(request.URL().Path, found.pattern)
	for name := range pathValues {
		request.Instance().SetPathValue(name, pathValues[name])
	}

	return found.handler
}

func (router *Router) Get(pattern string, handler func (*Request, *Response)) error {
	return router.add("GET", pattern, handler)
}

func (router *Router) Post(pattern string, handler func (*Request, *Response)) error {
	return router.add("POST", pattern, handler)
}

func routeParts(route string) ([]string, error) {
	route = strings.Trim(route, "/")
	route = strings.ReplaceAll(route, "//", "/")
	parts := strings.Split(route, "/")
	for i := range parts {
		esc, err := url.PathUnescape(parts[i])
		if err != nil {
			return []string{}, fmt.Errorf("%v - %w", ErrMessages[ERR_URL_DECODE_FAILURE], err)
		}
		parts[i] = esc
	}

	return parts, nil
}

func parsePath(path string, pattern string) map[string]string {
	paths, _ := routeParts(path)
	patterns, _ := routeParts(pattern)
	pathValue := make(map[string]string)
	for i := range patterns {
		if strings.HasPrefix(patterns[i], "{") && 
			strings.HasSuffix(patterns[i], "}") {
			key := strings.Trim(patterns[i], "{}")
			pathValue[key] = paths[i]
		}
	}

	return pathValue
}

func (router *Router) add(method string, pattern string, handler func (*Request, *Response)) error {
	// append method at the begining
	parts, err := routeParts(pattern)
	if err != nil {
		return err
	} 
	parts = append([]string {method}, parts...)
	// build route map
	node := router.routeMap 
	for _, part := range parts {
		var pathVal string
		if strings.HasPrefix(part, "{") {
			// path argument
			if ! strings.HasSuffix(part, "}") {
				return errors.New(ErrMessages[ERR_ARGUMENT_NOT_ENCLOSED])
			}
			pathVal = "*"
		} else {
			// fixed value
			pathVal = part
		}
		
		if _, ok := node.childs[pathVal]; !ok {
			node.childs[pathVal] = &RouteMap{
				name: pathVal,
				childs: make(map[string]*RouteMap),
			}
		}
		node = node.childs[pathVal]
	}
	// set handler
	node.handler = handler
	node.pattern = pattern

	return nil
}

func (router *Router) matchRoute(method string, pattern string) (*RouteMap, error) {
	parts, err := routeParts(pattern)
	if err != nil {
		return nil, err
	}

	routes, ok := router.routeMap.childs[method]
	if !ok {
		return nil, errors.New(ErrMessages[ERR_ROUTE_NOT_FOUND])
	}
	matches := routes.childs

	// childs := router.routeMap.childs
	for i, part := range parts {
		// loop through childs, take all whose value is * or match the current route part
		tmp := make(map[string]*RouteMap)
		for _, m := range matches {
			if m.name == "*" || m.name == part {
				if i == len(parts) - 1 {
					return m, nil
				}
				for _, ch := range m.childs {
					tmp[m.name + ch.name] = ch
				}
			}
		}
		// if we cannot find a child, return nil
		if len(tmp) == 0 && i != len(parts) - 1 {
			break
		}

		// proceed the next part
		matches = tmp
	}

	return nil, errors.New(ErrMessages[ERR_ROUTE_NOT_FOUND])
}
