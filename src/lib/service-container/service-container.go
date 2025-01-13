package container

import "sync"

type binding struct {
	singleton bool
	closure   func() any
}

type ServiceContainer struct {
	bindings  map[string]binding
	instances map[string]any
	mu        sync.Mutex
}

func NewContainer() *ServiceContainer {
	svContainer := ServiceContainer{}
	svContainer.bindings = make(map[string]binding)
	svContainer.instances = make(map[string]any)
	return &svContainer
}

func (container *ServiceContainer) Bind(name string, closure func() any) {
	container.mu.Lock()
	defer container.mu.Unlock()
	
	container.bindings[name] = binding{
		singleton: false,
		closure:   closure,
	}
}

func (container *ServiceContainer) BindSingleton(name string, closure func() any) {
	container.mu.Lock()
	defer container.mu.Unlock()

	container.bindings[name] = binding{
		singleton: true,
		closure:   closure,
	}
}

func (container *ServiceContainer) Resolve(name string) any {
	container.mu.Lock()
	defer container.mu.Unlock()

	binding, found := container.bindings[name]
	if !found {
		return nil
	}

	var inst any
	if !binding.singleton {
		inst = binding.closure()
	} else {
		if _, found := container.instances[name]; !found {
			container.instances[name] = binding.closure()
		}
		inst = container.instances[name]
	}

	return inst
}