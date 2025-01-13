package resthttp

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type Router struct {
	routeMap *RouteMap
}

func (router *Router) Func() func (http.ResponseWriter, *http.Request) {
	routerFunc := func (res http.ResponseWriter, req *http.Request) {
		request := Request { req: req }
		response := Response { writer: res }
		// get the handler
		match, err := router.matchRoute(req.Method, req.URL.Path)
		if err != nil {
			response.NotFound("The requested endpoint does not exist")
			return
		}
		// set path value
		pathValues := parsePath(req.URL.Path, match.pattern)
		for name := range pathValues {
			request.req.SetPathValue(name, pathValues[name])
		}
		// call handler
		match.handler.Handle(&request, &response)
	} 

	return routerFunc
}

func (router *Router) Get(pattern string, handler func(req *Request, res *Response)) error {
	return router.add("GET", pattern, &Handler{ Handle: handler })
}

func (router *Router) Post(pattern string, handler func(req *Request, res *Response)) error {
	return router.add("POST", pattern, &Handler{ Handle: handler })
}

func routeParts(route string) ([]string, error) {
	route = strings.Trim(route, "/")
	route = strings.ReplaceAll(route, "//", "/")
	parts := strings.Split(route, "/")
	for i := range parts {
		esc, err := url.PathUnescape(parts[i])
		if err != nil {
			return []string{}, err
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
			pathValue[patterns[i]] = paths[i]
		}
	}

	return pathValue
}

func (router *Router) add(method string, pattern string, handler *Handler) error {
	// append method at the begining
	parts, err := routeParts(pattern)
	if err != nil {
		return err
	} 
	parts = append(parts, method)
	// build route map
	node := router.routeMap 
	for _, part := range parts {
		var pathVal string
		if strings.HasPrefix(part, "{") {
			// path argument
			if ! strings.HasSuffix(part, "}") {
				return errors.New("path argument is not enclosed with \"{}\"")
			}
			pathVal = "*"
		} else {
			// fixed value
			pathVal = part
		}
		
		if _, ok := node.childs[pathVal]; !ok {
			node.childs[pathVal] = &RouteMap{
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
	parts = append(parts, method)

	childs := router.routeMap.childs
	for _, part := range parts {
		// loop through childs, take all whose value is * or match the current route part
		tmp := map[string]*RouteMap{}
		for value, ch := range childs {
			if value == "*" || value == part {
				tmp[value] = ch
			}
		}
		// if we cannot find a child, return nil
		if len(tmp) == 0 {
			return nil, errors.New("route does not exist")
		}
		// proceed the next part
		childs = tmp
	}

	// return the first match route and match method
	var result *RouteMap
	for key := range childs {
		result = childs[key]
		break
	}

	return result, nil
}
