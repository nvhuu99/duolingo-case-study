package router

import (
	"errors"
	"strings"
)

var (
	ErrAddEmptyRoute       = errors.New("empty route is now allowed")
	ErrArgumentNotEnclosed = errors.New("route argument is not enclosed")
	ErrRouteDuplication    = errors.New("adding a duplication route")
)

type Router struct {
	root *routeNode
}

func NewRouter() *Router {
	return &Router{
		root: &routeNode{},
	}
}

func (router *Router) Match(requestPath string) (bool, *RouteResult) {
	found, node := router.travel(requestPath)
	if !found {
		return false, nil
	}
	return true, newRouteResult(requestPath, node)
}

func (router *Router) Add(pattern string, handler any) error {
	parts := cleanPathArr(pattern)
	if len(parts) == 0 {
		return ErrAddEmptyRoute
	}
	travelNode := router.root
	for i := range parts {
		if strings.HasPrefix(parts[i], "{") && !strings.HasSuffix(parts[i], "}") {
			return ErrArgumentNotEnclosed
		}

		key := parts[i]
		if isPathArg(parts[i]) {
			key = "{DUMMY_ARG_NAME}"
		}

		partExists := false
		for _, child := range travelNode.childs {
			if !isPathArg(child.key) && child.key != key {
				continue
			}
			if i == len(parts)-1 {
				return ErrRouteDuplication
			}
			travelNode = child
			partExists = true
			break
		}

		if !partExists {
			newNode := &routeNode{key: key}
			travelNode.childs = append(travelNode.childs, newNode)
			travelNode = newNode
		}
	}

	travelNode.pattern = strings.Join(parts, "/")
	travelNode.handler = handler

	return nil
}

func (router *Router) travel(requestPath string) (bool, *routeNode) {
	parts := cleanPathArr(requestPath)
	if len(parts) == 0 {
		return false, nil
	}
	nodes := router.root.childs
	for i := range parts {
		matches := []*routeNode{}
		for _, node := range nodes {
			if node.key == parts[i] || isPathArg(node.key) {
				if i == len(parts)-1 && node.handler != nil && node.pattern != "" {
					return true, node
				}
				matches = append(matches, node.childs...)
			}
		}
		nodes = matches
	}
	return false, nil
}
