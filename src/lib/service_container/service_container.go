package container

import (
	"sync"
)

var (
	container *ServiceContainer
)

type ServiceContainer struct {
	bindings  map[string]binding
	instances map[string]any
	mu        sync.Mutex
}

func GetContainer() *ServiceContainer {
	if container == nil {
		container = new(ServiceContainer)
		container.bindings = make(map[string]binding)
		container.instances = make(map[string]any)
	}

	return container
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
	instance := closure()
	container.mu.Lock()
	container.bindings[name] = binding{singleton: true}
	container.instances[name] = instance
	container.mu.Unlock()
}

func (container *ServiceContainer) Resolve(name string) any {
	container.mu.Lock()
	binding, found := container.bindings[name]
	container.mu.Unlock()
	if !found {
		return nil
	}
	if !binding.singleton {
		return binding.closure()
	}
	container.mu.Lock()
	instance, found := container.instances[name]
	container.mu.Unlock()
	if !found {
		return nil
	} else {
		return instance
	}
}
