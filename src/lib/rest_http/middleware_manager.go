package rest_http

type MiddlewareManager struct {
	middlewares map[string][]Handler
}

func NewMiddlewareManager() *MiddlewareManager {
	manager := new(MiddlewareManager)
	manager.middlewares = make(map[string][]Handler)
	return manager
}

func (manager *MiddlewareManager) HasGroup(group string) bool {
	_, exist := manager.middlewares[group]
	return exist
}

func (manager *MiddlewareManager) Push(group string, middlewares ...Handler) {
	if !manager.HasGroup(group) {
		manager.middlewares[group] = []Handler{}
	}
	for _, handler := range middlewares {
		manager.middlewares[group] = append(manager.middlewares[group], handler)
	}
}

func (manager *MiddlewareManager) Handle(group string, request *Request, response *Response) {
	if !manager.HasGroup(group) {
		return
	}

	middlewares := manager.middlewares[group]
	if len(middlewares) == 0 {
		return
	}

	iterator := middlewares[0]
	for i := 1; i < len(middlewares); i++ {
		iterator.SetNext(middlewares[i])
		iterator = middlewares[i]
	}

	middlewares[0].Handle(request, response)
}
