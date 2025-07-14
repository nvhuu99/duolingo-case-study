package router

type routeNode struct {
	key    string
	childs []*routeNode

	pattern string
	handler any
}
